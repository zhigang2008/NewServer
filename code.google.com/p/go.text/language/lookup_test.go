// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import (
	"strings"
	"testing"
)

var strdata = []string{
	"aa  ",
	"aaa ",
	"aaaa",
	"aaab",
	"aab ",
	"ab  ",
	"ba  ",
	"xxxx",
}

func strtests() map[string]int {
	return map[string]int{
		"    ": 0,
		"a":    0,
		"aa":   0,
		"aaa":  4,
		"aa ":  0,
		"aaaa": 8,
		"aaab": 12,
		"aaax": 16,
		"b":    24,
		"ba":   24,
		"bbbb": 28,
	}
}

func TestSearch(t *testing.T) {
	for k, v := range strtests() {
		if i := search(strings.Join(strdata, ""), []byte(k)); i != v {
			t.Errorf("%s: found %d; want %d", k, i, v)
		}
	}
}

func TestIndex(t *testing.T) {
	strtests := strtests()
	strtests["    "] = -1
	strtests["aaax"] = -1
	strtests["bbbb"] = -1
	for k, v := range strtests {
		if i := index(strings.Join(strdata, ""), []byte(k)); i != v {
			t.Errorf("%s: found %d; want %d", k, i, v)
		}
	}
}

func b(s string) []byte {
	return []byte(s)
}

func TestFixCase(t *testing.T) {
	tests := []string{
		"aaaa", "AbCD", "abcd",
		"Zzzz", "AbCD", "Abcd",
		"Zzzz", "AbC", "Zzzz",
		"XXX", "ab ", "XXX",
		"XXX", "usd", "USD",
		"cmn", "AB ", "cmn",
		"gsw", "CMN", "cmn",
	}
	for i := 0; i+3 < len(tests); i += 3 {
		tt := tests[i:]
		buf := [4]byte{}
		b := buf[:copy(buf[:], tt[1])]
		res := fixCase(tt[0], b)
		if res && cmp(tt[2], b) != 0 || !res && tt[0] != tt[2] {
			t.Errorf("%s+%s: found %q; want %q", tt[0], tt[1], b, tt[2])
		}
	}
}

func TestLangID(t *testing.T) {
	tests := []struct {
		id, bcp47, iso3, norm string
		err                   error
	}{
		{id: "", bcp47: "und", iso3: "und", err: errSyntax},
		{id: "  ", bcp47: "und", iso3: "und", err: errSyntax},
		{id: "   ", bcp47: "und", iso3: "und", err: errSyntax},
		{id: "    ", bcp47: "und", iso3: "und", err: errSyntax},
		{id: "xxx", bcp47: "und", iso3: "und", err: mkErrInvalid([]byte("xxx"))},
		{id: "und", bcp47: "und", iso3: "und"},
		{id: "aju", bcp47: "aju", iso3: "aju", norm: "jrb"},
		{id: "jrb", bcp47: "jrb", iso3: "jrb"},
		{id: "es", bcp47: "es", iso3: "spa"},
		{id: "spa", bcp47: "es", iso3: "spa"},
		{id: "ji", bcp47: "ji", iso3: "yid-", norm: "yi"},
		{id: "jw", bcp47: "jw", iso3: "jav-", norm: "jv"},
		{id: "ar", bcp47: "ar", iso3: "ara"},
		{id: "kw", bcp47: "kw", iso3: "cor"},
		{id: "arb", bcp47: "arb", iso3: "arb", norm: "ar"},
		{id: "ar", bcp47: "ar", iso3: "ara"},
		{id: "kur", bcp47: "ku", iso3: "kur"},
		{id: "nl", bcp47: "nl", iso3: "nld"},
		{id: "NL", bcp47: "nl", iso3: "nld"},
		{id: "gsw", bcp47: "gsw", iso3: "gsw"},
		{id: "gSW", bcp47: "gsw", iso3: "gsw"},
		{id: "und", bcp47: "und", iso3: "und"},
		{id: "sh", bcp47: "sh", iso3: "hbs", norm: "sh"},
		{id: "hbs", bcp47: "sh", iso3: "hbs", norm: "sh"},
		{id: "no", bcp47: "no", iso3: "nor", norm: "no"},
		{id: "nor", bcp47: "no", iso3: "nor", norm: "no"},
		{id: "cmn", bcp47: "cmn", iso3: "cmn", norm: "zh"},
	}
	for i, tt := range tests {
		want, err := getLangID(b(tt.id))
		if err != tt.err {
			t.Errorf("%d:err(%s): found %q; want %q", i, tt.id, err, tt.err)
		}
		if err != nil {
			continue
		}
		if id, _ := getLangISO2(b(tt.bcp47)); len(tt.bcp47) == 2 && want != id {
			t.Errorf("%d:getISO2(%s): found %v; want %v", i, tt.bcp47, id, want)
		}
		if len(tt.iso3) == 3 {
			if id, _ := getLangISO3(b(tt.iso3)); want != id {
				t.Errorf("%d:getISO3(%s): found %q; want %q", i, tt.iso3, id, want)
			}
			if id, _ := getLangID(b(tt.iso3)); want != id {
				t.Errorf("%d:getID3(%s): found %v; want %v", i, tt.iso3, id, want)
			}
		}
		norm := want
		if tt.norm != "" {
			norm, _ = getLangID(b(tt.norm))
		}
		id := normLang(langOldMap[:], want)
		id = normLang(langMacroMap[:], id)
		if id != norm {
			t.Errorf("%d:norm(%s): found %v; want %v", i, tt.id, id, norm)
		}
		if id := want.String(); tt.bcp47 != id {
			t.Errorf("%d:String(): found %s; want %s", i, id, tt.bcp47)
		}
		if id := want.ISO3(); tt.iso3[:3] != id {
			t.Errorf("%d:iso3(): found %s; want %s", i, id, tt.iso3[:3])
		}
	}
}

func TestRegionID(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"_  ", ""},
		{"_000", ""},
		{"419", "419"},
		{"AA", "AA"},
		{"ATF", "TF"},
		{"HV", "HV"},
		{"CT", "CT"},
		{"DY", "DY"},
		{"IC", "IC"},
		{"FQ", "FQ"},
		{"JT", "JT"},
		{"ZZ", "ZZ"},
		{"EU", "EU"},
		{"QO", "QO"},
		{"FX", "FX"},
	}
	for i, tt := range tests {
		if tt.in[0] == '_' {
			id := tt.in[1:]
			if _, err := getRegionID(b(id)); err == nil {
				t.Errorf("%d:err(%s): found nil; want error", i, id)
			}
			continue
		}
		want, _ := getRegionID(b(tt.in))
		if s := want.String(); s != tt.out {
			t.Errorf("%d:%s: found %q; want %q", i, tt.in, s, tt.out)
		}
		if len(tt.in) == 2 {
			want, _ := getRegionISO2(b(tt.in))
			if s := want.String(); s != tt.out {
				t.Errorf("%d:getISO2(%s): found %q; want %q", i, tt.in, s, tt.out)
			}
		}
	}
}

func TestRegionISO3(t *testing.T) {
	tests := []struct {
		from, iso3, to string
	}{
		{"  ", "", ""},
		{"000", "", ""},
		{"AA", "AAA", ""},
		{"CT", "CTE", ""},
		{"DY", "DHY", ""},
		{"EU", "QUU", ""},
		{"HV", "HVO", ""},
		{"IC", "", ""},
		{"JT", "JTN", ""},
		{"PZ", "PCZ", ""},
		{"QU", "QUU", "EU"},
		{"QO", "QOO", ""},
		{"YD", "YMD", ""},
		{"FQ", "ATF", "TF"},
		{"TF", "ATF", ""},
		{"FX", "FXX", ""},
		{"ZZ", "ZZZ", ""},
		{"419", "", ""},
	}
	for _, tt := range tests {
		r, _ := getRegionID(b(tt.from))
		if s := r.ISO3(); s != tt.iso3 {
			t.Errorf("iso3(%q): found %q; want %q", tt.from, s, tt.iso3)
		}
		if tt.iso3 == "" {
			continue
		}
		want := tt.to
		if tt.to == "" {
			want = tt.from
		}
		r, _ = getRegionID(b(want))
		if id, _ := getRegionISO3(b(tt.iso3)); id != r {
			t.Errorf("%s: found %q; want %q", tt.iso3, id, want)
		}
	}
}

func TestRegionM49(t *testing.T) {
	fromTests := []struct {
		m49 int
		id  string
	}{
		{0, ""},
		{-1, ""},
		{1000, ""},
		{10000, ""},

		{001, "001"},
		{104, "MM"},
		{180, "CD"},
		{230, "ET"},
		{231, "ET"},
		{249, "FX"},
		{250, "FR"},
		{276, "DE"},
		{278, "DD"},
		{280, "DE"},
		{419, "419"},
		{626, "TL"},
		{736, "SD"},
		{840, "US"},
		{854, "BF"},
		{891, "CS"},
		{899, ""},
		{958, "AA"},
		{966, "QT"},
		{967, "EU"},
		{999, "ZZ"},
	}
	for _, tt := range fromTests {
		id, err := getRegionM49(tt.m49)
		if want, have := err != nil, tt.id == ""; want != have {
			t.Errorf("error(%d): have %v; want %v", tt.m49, have, want)
			continue
		}
		r, _ := getRegionID(b(tt.id))
		if r != id {
			t.Errorf("region(%d): have %s; want %s", tt.m49, id, r)
		}
	}

	toTests := []struct {
		m49 int
		id  string
	}{
		{0, "000"},
		{0, "IC"}, // Some codes don't have an ID

		{001, "001"},
		{104, "MM"},
		{104, "BU"},
		{180, "CD"},
		{180, "ZR"},
		{231, "ET"},
		{250, "FR"},
		{249, "FX"},
		{276, "DE"},
		{278, "DD"},
		{419, "419"},
		{626, "TL"},
		{626, "TP"},
		{729, "SD"},
		{826, "GB"},
		{840, "US"},
		{854, "BF"},
		{891, "YU"},
		{891, "CS"},
		{958, "AA"},
		{966, "QT"},
		{967, "EU"},
		{967, "QU"},
		{999, "ZZ"},
		// For codes that don't have an M49 code use the replacement value,
		// if available.
		{854, "HV"}, // maps to Burkino Faso
	}
	for _, tt := range toTests {
		r, _ := getRegionID(b(tt.id))
		if r.M49() != tt.m49 {
			t.Errorf("m49(%q): have %d; want %d", tt.id, r.M49(), tt.m49)
		}
	}
}

func TestRegionDeprecation(t *testing.T) {
	tests := []struct{ in, out string }{
		{"BU", "MM"},
		{"BUR", "MM"},
		{"CT", "KI"},
		{"DD", "DE"},
		{"DDR", "DE"},
		{"DY", "BJ"},
		{"FX", "FR"},
		{"HV", "BF"},
		{"JT", "UM"},
		{"MI", "UM"},
		{"NH", "VU"},
		{"NQ", "AQ"},
		{"PU", "UM"},
		{"PZ", "PA"},
		{"QU", "EU"},
		{"RH", "ZW"},
		{"TP", "TL"},
		{"UK", "GB"},
		{"VD", "VN"},
		{"WK", "UM"},
		{"YD", "YE"},
		{"NL", "NL"},
	}
	for _, tt := range tests {
		rIn, _ := getRegionID([]byte(tt.in))
		rOut, _ := getRegionISO2([]byte(tt.out))
		r := normRegion(rIn)
		if rOut == rIn && r != 0 {
			t.Errorf("%s: was %q; want %q", tt.in, r, tt.in)
		}
		if rOut != rIn && r != rOut {
			t.Errorf("%s: was %q; want %q", tt.in, r, tt.out)
		}

	}
}

func TestGetScriptID(t *testing.T) {
	idx := "0000BbbbDdddEeeeZzzz\xff\xff\xff\xff"
	tests := []struct {
		in  string
		out scriptID
	}{
		{"    ", 0},
		{"      ", 0},
		{"  ", 0},
		{"", 0},
		{"Aaaa", 0},
		{"Bbbb", 1},
		{"Dddd", 2},
		{"dddd", 2},
		{"dDDD", 2},
		{"Eeee", 3},
		{"Zzzz", 4},
	}
	for i, tt := range tests {
		if id, err := getScriptID(idx, b(tt.in)); id != tt.out {
			t.Errorf("%d:%s: found %d; want %d", i, tt.in, id, tt.out)
		} else if id == 0 && err == nil {
			t.Errorf("%d:%s: no error; expected one", i, tt.in)
		}
	}
}

func TestCurrency(t *testing.T) {
	curInfo := func(round, dec int) string {
		return string(round<<2 + dec)
	}
	idx := strings.Join([]string{
		"   \x00",
		"BBB" + curInfo(5, 2),
		"DDD\x00",
		"XXX\x00",
		"ZZZ\x00",
		"\xff\xff\xff\xff",
	}, "")
	tests := []struct {
		in         string
		out        currencyID
		round, dec int
	}{
		{"   ", 0, 0, 0},
		{"     ", 0, 0, 0},
		{" ", 0, 0, 0},
		{"", 0, 0, 0},
		{"BBB", 1, 5, 2},
		{"DDD", 2, 0, 0},
		{"dDd", 2, 0, 0},
		{"ddd", 2, 0, 0},
		{"XXX", 3, 0, 0},
		{"Zzz", 4, 0, 0},
	}
	for i, tt := range tests {
		id, err := getCurrencyID(idx, b(tt.in))
		if id != tt.out {
			t.Errorf("%d:%s: found %d; want %d", i, tt.in, id, tt.out)
		} else if tt.out == 0 && err == nil {
			t.Errorf("%d:%s: no error; expected one", i, tt.in)
		}
		if id > 0 {
			if d := decimals(idx, id); d != tt.dec {
				t.Errorf("%d:dec(%s): found %d; want %d", i, tt.in, d, tt.dec)
			}
			if d := round(idx, id); d != tt.round {
				t.Errorf("%d:round(%s): found %d; want %d", i, tt.in, d, tt.round)
			}
		}
	}
}

func TestIsPrivateUse(t *testing.T) {
	type test struct {
		s       string
		private bool
	}
	tests := []test{
		{"en", false},
		{"und", false},
		{"pzn", false},
		{"qaa", true},
		{"qtz", true},
		{"qua", false},
	}
	for i, tt := range tests {
		x, _ := getLangID([]byte(tt.s))
		if b := x.IsPrivateUse(); b != tt.private {
			t.Errorf("%d: langID.IsPrivateUse(%s) was %v; want %v", i, tt.s, b, tt.private)
		}
	}
	tests = []test{
		{"001", false},
		{"419", false},
		{"899", false},
		{"900", false},
		{"957", false},
		{"958", true},
		{"AA", true},
		{"AC", false},
		{"EU", true}, // CLDR grouping
		{"QO", true}, // CLDR grouping
		{"QA", false},
		{"QM", true},
		{"QZ", true},
		{"XA", true},
		{"XZ", true},
		{"ZW", false},
		{"ZZ", true},
	}
	for i, tt := range tests {
		x, _ := getRegionID([]byte(tt.s))
		if b := x.IsPrivateUse(); b != tt.private {
			t.Errorf("%d: regionID.IsPrivateUse(%s) was %v; want %v", i, tt.s, b, tt.private)
		}
	}
	tests = []test{
		{"Latn", false},
		{"Laaa", false}, // invalid
		{"Qaaa", true},
		{"Qabx", true},
		{"Qaby", false},
		{"Zyyy", false},
		{"Zzzz", false},
	}
	for i, tt := range tests {
		x, _ := getScriptID(script, []byte(tt.s))
		if b := x.IsPrivateUse(); b != tt.private {
			t.Errorf("%d: scriptID.IsPrivateUse(%s) was %v; want %v", i, tt.s, b, tt.private)
		}
	}
}
