// Copyright 2024 Mateusz Butkiewicz
//
// Original author: Andreas Wagner
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package langnegmatcher

import (
	"errors"
	"fmt"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"net/http"
	"strconv"
	"strings"
)

type Config struct {
	// List of language codes to match against ([IETF RFC 7231, section 5.3.5](https://datatracker.ietf.org/doc/html/rfc7231#section-5.3.5)). Default: Empty list
	MatchLanguages []string
	// Indicator to include closest to full locale (e.g. en-US) or only language (e.g. en for en-US). Default: false
	FullLocale bool
	// Variable name (will be prefixed with `lanneg_`) to hold result of language negotiation. Default: ""
	VarLanguage string
	// Hardcoded value used if matcher do not match any value. VarLanguage will be set with it. Default: ""
	FallbackValue string
}

func (c *Config) UnmarshalFromCaddy(d *caddyfile.Dispenser) error {
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "match_languages":
				c.MatchLanguages = append(c.MatchLanguages, d.RemainingArgs()...)
			case "full_locale":
				d.Next()
				val := d.Val()
				boolVal, err := strconv.ParseBool(val)
				if err != nil {
					return err
				}
				c.FullLocale = boolVal
			case "var_language":
				d.Next()
				c.VarLanguage = d.Val()
			case "fallback_value":
				d.Next()
				c.FallbackValue = d.Val()
			}
		}
	}
	return nil
}

// Matcher matches requests by comparing results of a
// content negotiation (specifically language) process to a (list of) value(s).
//
// Languages to match the request against given values - and at least one of them MUST
// be specified.
//
// COMPATIBILITY NOTE: This module is still experimental and is not
// subject to Caddy's compatibility guarantee.
type Matcher struct {
	Config Config

	LanguageMatcher language.Matcher
	logger          *zap.Logger
}

func init() {
	caddy.RegisterModule(&Matcher{})
}

// CaddyModule returns the Caddy module information.
func (*Matcher) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.langneg",
		New: func() caddy.Module { return new(Matcher) },
	}
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Matcher) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	cfg := &Config{}
	err := cfg.UnmarshalFromCaddy(d)
	if err != nil {
		m.logger.Error("error unmarshalling caddy into config", zap.Error(err))
		return err
	}
	m.Config = *cfg
	return nil
}

// Provision sets up the module.
func (m *Matcher) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger()
	var MatchTLanguages []language.Tag
	MatchTLanguages = append(MatchTLanguages, language.Und)
	for _, l := range m.Config.MatchLanguages {
		MatchTLanguages = append(MatchTLanguages, language.Make(l))
	}
	m.LanguageMatcher = language.NewMatcher(MatchTLanguages)
	return nil
}

// Validate validates that the module has a usable config.
func (m *Matcher) Validate() error {
	if len(m.Config.MatchLanguages) == 0 && len(m.Config.VarLanguage) > 0 {
		return errors.New("you cannot specify a variable to store content negotiation results (for languages) if you don't also specify what languages are offered. (Use '*' to work around this constraint.)")
	}
	return nil
}

// Match returns true if the request matches all requirements. If fails and fallback value is set returns true and uses fallback value.
func (m *Matcher) Match(r *http.Request) bool {

	languageMatch, locale := false, ""
	if len(m.Config.MatchLanguages) == 0 {
		languageMatch = true
	} else {
		languageMatch, locale = m.matchLanguage(r)
		if languageMatch && len(m.Config.VarLanguage) > 0 {
			m.logger.Debug("matched value", zap.String(m.Config.VarLanguage, locale))
			caddyhttp.SetVar(r.Context(), "langneg_"+m.Config.VarLanguage, locale)
		} else if len(m.Config.FallbackValue) > 0 && len(m.Config.VarLanguage) > 0 {
			m.logger.Debug("using fallback value", zap.String(m.Config.VarLanguage, m.Config.FallbackValue))
			caddyhttp.SetVar(r.Context(), "langneg_"+m.Config.VarLanguage, m.Config.FallbackValue)
			return true
		}
	}

	return languageMatch
}

func (m *Matcher) matchLanguage(r *http.Request) (bool, string) {
	match, result := false, ""
	headerValue := r.Header.Get("Accept-Language")
	m.logger.Debug("Header Accept-Language", zap.String("headerValue", headerValue))
	m.logger.Debug("Match language values", zap.Strings("matchLanguages", m.Config.MatchLanguages))

	tag, idx := language.MatchStrings(m.LanguageMatcher, headerValue)
	fmt.Print(idx)
	match = !tag.IsRoot()
	if match {
		if m.Config.FullLocale {
			var res []string
			b, bc := tag.Base()
			r, rc := tag.Region()
			s, sc := tag.Script()

			if bc == language.Exact {
				res = append(res, b.String())
			}

			if rc == language.Exact {
				res = append(res, r.String())
			}

			if sc == language.Exact {
				res = append(res, s.String())
			}
			result = strings.Join(res, "-")
		} else {
			b, _ := tag.Base()
			result = b.String()
		}
	} else {
		result = ""
	}
	return match, result
}

// Interface guards
var (
	_ caddyhttp.RequestMatcher = (*Matcher)(nil)
	_ caddyfile.Unmarshaler    = (*Matcher)(nil)
	_ caddy.Provisioner        = (*Matcher)(nil)
	_ caddy.Validator          = (*Matcher)(nil)
)
