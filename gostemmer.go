package gostemmer

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"
)

type Dict struct {
	class string
	lemma string
}

type Result struct {
	Count int
	Roots map[string]map[string]string
}

type Affix struct {
	a, b interface{}
}

type Prefix struct {
	a, b, c interface{}
}

func addRoot(roots map[string]map[string]string, is_suffix bool, pattern string, variant string) bool {
	pattern = "^" + pattern + "$"

	add_to_root := map[string]map[string]string{}
	for lemma, attrib := range roots {
		match_regex, _ := regexp.Compile(pattern)
		matches_all := match_regex.FindStringSubmatch(lemma)
		if len(matches_all) > 0 {
			matches := matches_all[1:len(matches_all)]
			if len(matches) > 0 {
				new_lemma := ""
				new_affix := ""

				var affix_index int
				if is_suffix {
					affix_index = 1
				} else {
					affix_index = 0
				}

				for x := range matches {
					if x != affix_index {
						new_lemma = new_lemma + matches[x]
					}
				}

				if len(variant) > 0 {
					new_lemma = variant + new_lemma
				}

				if is_suffix {
					new_affix = new_affix + "-"
				}
				new_affix = new_affix + matches[affix_index]
				if !is_suffix {
					new_affix = new_affix + "-"
				}

				if (reflect.ValueOf(attrib).Kind()) == reflect.Map {
					if len(attrib["affixes"]) == 0 {
						new_affix = attrib["affixes"] + new_affix
					} else {
						new_affix = attrib["affixes"] + ", " + new_affix
					}
				}

				add_to_root[new_lemma] = map[string]string{"affixes": new_affix}
			}
		}
	}

	for key, item := range add_to_root {
		roots[key] = item
	}

	return true

}

func checkArray(key string, arr []string) bool {
	for _, val := range arr {
		if val == key {
			return true
		}
	}
	return false
}

func stemWord(word string, options map[string]bool, affixes []Affix,
	prefixes []Prefix, disallowed_confixes [][]string, allomorphs map[string][]string,
	dicts map[string]Dict) map[string]map[string]string {
	word = strings.TrimSpace(word)
	roots := map[string]map[string]string{
		word: map[string]string{
			"affixes": "",
		},
	}

	if strings.ContainsAny(word, "-") {
		dash_parts := strings.Split(word, "-")
		for _, dash_part := range dash_parts {
			roots[dash_part] = map[string]string{
				"affixes": "",
			}
		}
	}

	for _, group := range affixes {
		is_suffix := group.a.(bool)
		affix_list := group.b.([]string)
		for _, affix := range affix_list {
			if is_suffix {
				pattern := "(.+)(" + affix + ")"
				addRoot(roots, is_suffix, pattern, "")
			} else {
				pattern := "(" + affix + ")(.+)"
				addRoot(roots, is_suffix, pattern, "")
			}
		}
	}

	for i := 0; i < 4; i++ {
		for _, rule := range prefixes {
			addRoot(roots, rule.a.(bool), rule.b.(string), rule.c.(string))
		}
	}

	to_delete := []string{}

	for lemma, attrib := range roots {
		if _, ok := dicts[lemma]; !ok {
			to_delete = append(to_delete, lemma)
			continue
		}

		if !options["STRICT_CONFIX"] {
			continue
		}

		new_affixes := attrib["affixes"]
		for _, pair := range disallowed_confixes {
			prefix := pair[0]
			suffix := pair[1]
			prefix_key := prefix[0:2]
			if _, ok := allomorphs[prefix_key]; ok {
				for _, allomorph := range allomorphs[prefix_key] {
					if checkArray(allomorph, strings.Split(new_affixes, ",")) && checkArray(suffix, strings.Split(new_affixes, ",")) {
						to_delete = append(to_delete, lemma)
					}
				}
			} else if checkArray(prefix, strings.Split(new_affixes, ",")) && checkArray(suffix, strings.Split(new_affixes, ",")) {
				to_delete = append(to_delete, lemma)
			}
		}

	}

	for _, del := range to_delete {
		delete(roots, del)
	}

	for lemma, attrib := range roots {
		attrib["lemma"] = dicts[lemma].lemma
		attrib["class"] = dicts[lemma].class
		var type_ string
		for _, aff := range strings.Split(attrib["affixes"], ",") {
			if len(aff) > 0 {
				if aff[0:1] == "-" {
					type_ = "suffixes"
				} else {
					type_ = "prefixes"
				}
				attrib[type_] = aff
			}

		}

		roots[lemma] = attrib
	}

	return roots
}

func stem(query string, options map[string]bool, affixes []Affix,
	prefixes []Prefix, disallowed_confixes [][]string, allomorphs map[string][]string,
	dicts map[string]Dict) map[string]*Result {
	words := map[string]map[string]int{}
	raw := strings.Split(query, " ")
	reg, _ := regexp.Compile("^\\d+$")

	for _, r := range raw {
		if options["NO_DIGIT_ONLY"] {
			if len(reg.FindAllString(r, -1)) > 1 {
				continue
			}
		}
		words[r] = map[string]int{
			"count": 0,
		}
		words[r]["count"] = words[r]["count"] + 1
	}

	final_result := map[string]*Result{}
	for key, word := range words {
		stemmed := stemWord(key, options, affixes, prefixes, disallowed_confixes, allomorphs, dicts)
		final_result[key] = new(Result)
		final_result[key].Count = word["count"]
		final_result[key].Roots = stemmed

		if len(stemmed) == 0 && options["NO_NO_MATCH"] {
			delete(final_result, key)
		}
	}

	return final_result
}

func StemWord(word string, kamus string) map[string]*Result {
	// Set Kamus
	dictionary, err := os.Open(kamus)
	if err != nil {
		fmt.Println("Kamus tidak ditemukan!")
		os.Exit(0)
	}
	defer dictionary.Close()

	dicts := map[string]Dict{}

	scanner := bufio.NewScanner(dictionary)
	for scanner.Scan() {
		stringClass := strings.Split(scanner.Text(), "\t")[0]
		stringLemma := strings.Split(scanner.Text(), "\t")[1]

		tempDict := new(Dict)
		tempDict.class = stringClass
		tempDict.lemma = stringLemma

		dicts[stringLemma] = *tempDict
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Terjadi error pada kamus!")
		os.Exit(0)
	}

	// Set option
	options := map[string]bool{
		"SORT_INSTANCE": false,
		"NO_NO_MATCH":   false,
		"NO_DIGIT_ONLY": true,
		"STRICT_CONFIX": false,
	}

	// Set rules
	VOWEL := "a|i|u|e|o"
	CONSONANT := "b|c|d|f|g|h|j|k|l|m|n|p|q|r|s|t|v|w|x|y|z"
	ANY := VOWEL + "|" + CONSONANT

	affixes := []Affix{}
	affixes = append(affixes, Affix{true, []string{"kah", "lah", "tah", "pun"}})
	affixes = append(affixes, Affix{true, []string{"mu", "ku", "nya"}})
	affixes = append(affixes, Affix{false, []string{"ku", "kau"}})
	affixes = append(affixes, Affix{true, []string{"i", "kan", "an"}})

	prefixes := []Prefix{}
	prefixes = append(prefixes, Prefix{false, "(di|ke|se)(" + ANY + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(ber|ter)(" + ANY + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(be|te)(r)(" + ANY + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(be|te)(" + CONSONANT + ")(" + ANY + ")(er)(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(bel|pel)(ajar|unjur)", ""})
	prefixes = append(prefixes, Prefix{false, "(me|pe)(l|m|n|r|w|y)(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(mem|pem)(b|f|v)(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(men|pen)(c|d|j|z)(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(meng|peng)(g|h|q|x)(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(meng|peng)(" + VOWEL + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(mem|pem)(" + VOWEL + ")(.+)", "p"})
	prefixes = append(prefixes, Prefix{false, "(men|pen)(" + VOWEL + ")(.+)", "t"})
	prefixes = append(prefixes, Prefix{false, "(meng|peng)(" + VOWEL + ")(.+)", "k"})
	prefixes = append(prefixes, Prefix{false, "(meny|peny)(" + VOWEL + ")(.+)", "s"})
	prefixes = append(prefixes, Prefix{false, "(mem)(p)(" + CONSONANT + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(pem)(" + CONSONANT + ")(.+)", "p"})
	prefixes = append(prefixes, Prefix{false, "(men|pen)(t)(" + CONSONANT + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(meng|peng)(k)(" + CONSONANT + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(men|pen)(s)(" + CONSONANT + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(menge|penge)(" + CONSONANT + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(mempe)(r)(" + VOWEL + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(memper)(" + ANY + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(pe)(" + ANY + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(per)(" + ANY + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(pel)(" + CONSONANT + ")(.+)", ""})
	prefixes = append(prefixes, Prefix{false, "(mem)(punya)", ""})
	prefixes = append(prefixes, Prefix{false, "(pen)(yair)", ""})

	disallowed_confixes := [][]string{}
	disallowed_confixes = append(disallowed_confixes, []string{"ber-", "-i"})
	disallowed_confixes = append(disallowed_confixes, []string{"ke-", "-i"})
	disallowed_confixes = append(disallowed_confixes, []string{"pe-", "-kan"})
	disallowed_confixes = append(disallowed_confixes, []string{"meng-", "-an"})
	disallowed_confixes = append(disallowed_confixes, []string{"ter-", "-an"})
	disallowed_confixes = append(disallowed_confixes, []string{"ku-", "-an"})

	allomorphs := map[string][]string{
		"be": []string{"be-", "ber-", "bel-"},
		"te": []string{"te-", "ter-", "tel-"},
		"pe": []string{
			"pe-",
			"per-",
			"pel-",
			"pen-",
			"pem-",
			"peng-",
			"peny-",
			"penge-",
		},
		"me": []string{
			"me-",
			"men-",
			"mem-",
			"meng-",
			"meny-",
			"menge-",
		},
	}

	rooted := stem(word, options, affixes, prefixes, disallowed_confixes, allomorphs, dicts)
	return rooted
}
