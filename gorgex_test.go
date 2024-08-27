package gorgex

import (
	"fmt"
	"testing"
)

func TestSimpleNfa(t *testing.T) {
	data := []struct {
		match_regex string
		to_match    string
		validity    bool
	}{
		{match_regex: `abc`, to_match: "abc", validity: true},
		{match_regex: `this should also match`, to_match: "this should also match", validity: true},
		{match_regex: `123`, to_match: "123", validity: true},
	}

	for _, instance := range data {
		ctx := parse(instance.match_regex)
		nfa := toNfa(ctx)
		t.Run(fmt.Sprintf("Test: '%s'", instance.to_match), func(t *testing.T) {
			result := nfa.check(instance.to_match, -1)
			if result != instance.validity {
				t.Logf("Expect: %t, got %t\n", instance.validity, result)
				t.Fail()
			}
		})
	}
}

func TestBracketNfa(t *testing.T) {
	data := []struct {
		match_regex string
		to_match    string
		validity    bool
	}{
		{match_regex: `ab[cd]`, to_match: "abc", validity: true},
		{match_regex: `ab[cd]`, to_match: "abd", validity: true},
		{match_regex: `ab[cd]`, to_match: "abr", validity: false},
		{match_regex: `ab[c-f]`, to_match: "abd", validity: true},
		{match_regex: `ab[cd_]`, to_match: "ab_", validity: true},
		{match_regex: `ab[c-f_.]+`, to_match: "abcdef_.", validity: true},
	}

	for _, instance := range data {
		ctx := parse(instance.match_regex)
		nfa := toNfa(ctx)
		t.Run(fmt.Sprintf("Test: '%s'", instance.to_match), func(t *testing.T) {
			result := nfa.check(instance.to_match, -1)
			if result != instance.validity {
				t.Logf("Expect: %t, got %t\n", instance.validity, result)
				t.Fail()
			}
		})
	}
}

func TestRepeatNfa(t *testing.T) {
	data := []struct {
		match_regex string
		to_match    string
		validity    bool
	}{
		{match_regex: `ab?c`, to_match: "ac", validity: true},
		{match_regex: `ab?c`, to_match: "abc", validity: true},
		{match_regex: `ab+c`, to_match: "abc", validity: true},
		{match_regex: `ab+c`, to_match: "ac", validity: false},
		{match_regex: `ab+c`, to_match: "abbc", validity: true},
		{match_regex: `ab*c`, to_match: "ac", validity: true},
		{match_regex: `ab*c`, to_match: "abc", validity: true},
		{match_regex: `ab*c`, to_match: "abbbc", validity: true},
	}

	for _, instance := range data {
		ctx := parse(instance.match_regex)
		nfa := toNfa(ctx)
		t.Run(fmt.Sprintf("Test: '%s'", instance.to_match), func(t *testing.T) {
			result := nfa.check(instance.to_match, -1)
			if result != instance.validity {
				t.Logf("Expect: %t, got %t\n", instance.validity, result)
				t.Fail()
			}
		})
	}
}

func TestComboNfa(t *testing.T) {
	data := []struct {
		match_regex string
		to_match    string
		validity    bool
	}{
		{match_regex: `[a-zA-Z][a-zA-Z0-9_.]+`, to_match: "aqz", validity: true},
		{match_regex: `[a-zA-Z][a-zA-Z0-9_.]+@`, to_match: "john_smith.55@", validity: true},
		{match_regex: `[a-zA-Z][a-zA-Z0-9_.]+@[a-zA-Z0-9]+.`, to_match: "johnsmith@gmail.", validity: true},
		{match_regex: `[a-zA-Z][a-zA-Z0-9_.]+@[a-zA-Z0-9]+.[a-zA-Z]`, to_match: "johnsmith@gmail.c", validity: true},
		{match_regex: `[a-zA-Z][a-zA-Z0-9_.]+@[a-zA-Z0-9]+.[a-zA-Z]{2,}`, to_match: "johnsmith@gmail.co", validity: true},
	}

	for _, instance := range data {
		ctx := parse(instance.match_regex)
		nfa := toNfa(ctx)
		t.Run(fmt.Sprintf("Test: '%s'", instance.to_match), func(t *testing.T) {
			result := nfa.check(instance.to_match, -1)
			if result != instance.validity {
				t.Logf("Expect: %t, got %t\n", instance.validity, result)
				t.Fail()
			}
		})
	}
}

func TestEmailNfa(t *testing.T) {
	data := []struct {
		email    string
		validity bool
	}{
		{email: "valid_email@example.com", validity: true},
		{email: "john.doe@email.com", validity: true},
		{email: "user_name@email.org", validity: true},
		{email: "support@email.io", validity: true},
		{email: "contact@123.com", validity: true},
		{email: "sales@email.biz", validity: true},
		{email: "test_email@email.test", validity: true},
		{email: "random.email@email.xyz", validity: true},
		{email: "user@domain12345.com", validity: true},
		{email: "user@12345domain.com", validity: true},
		// invalid when compared against our regex
		{email: "alice.smith123@email.co.uk", validity: false},
		{email: "invalid.email@", validity: false},
		{email: ".invalid@email.com", validity: false},
		{email: "email@invalid..com", validity: false},
		{email: "user@-invalid.com", validity: false},
		{email: "user@invalid-.com", validity: false},
		{email: "user@in valid.com", validity: false},
		{email: "user@.com", validity: false},
		{email: "user@.co", validity: false},
		{email: "user@domain.c", validity: false},
		{email: "user@domain.1a", validity: false},
		{email: "user@domain.c0m", validity: false},
		{email: "user@domain..com", validity: false},
		{email: "user@.email.com", validity: false},
		{email: "user@emai.l.com", validity: false},
		{email: "user@e_mail.com", validity: false},
		{email: "user@e+mail.com", validity: false},
		{email: "user@e^mail.com", validity: false},
		{email: "user@e*mail.com", validity: false},
		{email: "user@e.mail.com", validity: false},
		{email: "user@e_mail.net", validity: false},
		{email: "user@sub.domain.com", validity: false},
		{email: "user@sub-domain.com", validity: false},
		{email: "user@sub.domain12345.com", validity: false},
		{email: "user@sub.domain-12345.com", validity: false},
		{email: "user@-sub.domain.com", validity: false},
		{email: "user@sub-.domain.com", validity: false},
		{email: "user@domain-.com", validity: false},
		{email: "user@sub.domain.c0m", validity: false},
		{email: "user@sub.domain.c", validity: false},
		{email: "user@sub.domain.1a", validity: false},
		{email: "user@sub.domain.c0m", validity: false},
		{email: "user@sub.domain..com", validity: false},
		{email: "user@sub.domain.c0m", validity: false},
		{email: "user@sub.domain..com", validity: false},
		{email: "user@sub.domain.c0m", validity: false},
	}

	ctx := parse(`[a-zA-Z][a-zA-Z0-9_.]+@[a-zA-Z0-9]+.[a-zA-Z]{2,}`)
	nfa := toNfa(ctx)

	for _, instance := range data {
		t.Run(fmt.Sprintf("Test: '%s'", instance.email), func(t *testing.T) {
			result := nfa.check(instance.email, -1)
			if result != instance.validity {
				t.Logf("Expect: %t, got %t\n", instance.validity, result)
				t.Fail()
			}
		})
	}
}
