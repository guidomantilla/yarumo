package utils

import (
	"reflect"
	"slices"
	"testing"
)

// helper: compare slices as multisets (order-insensitive)
func sameMultiset[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[T]int, len(a))
	for _, v := range a {
		m[v]++
	}

	for _, v := range b {
		m[v]--
	}

	for _, c := range m {
		if c != 0 {
			return false
		}
	}

	return true
}

// --- Predicate functions ---

func TestTernary(t *testing.T) {
	t.Parallel()

	t.Run("true condition returns trueValue", func(t *testing.T) {
		t.Parallel()

		got := Ternary(true, 1, 2)
		if got != 1 {
			t.Fatalf("got %d, want 1", got)
		}
	})

	t.Run("false condition returns falseValue", func(t *testing.T) {
		t.Parallel()

		got := Ternary(false, 1, 2)
		if got != 2 {
			t.Fatalf("got %d, want 2", got)
		}
	})

	t.Run("with string values", func(t *testing.T) {
		t.Parallel()

		got := Ternary(true, "yes", "no")
		if got != "yes" {
			t.Fatalf("got %q, want %q", got, "yes")
		}
	})
}

func TestCoalesce(t *testing.T) {
	t.Parallel()

	t.Run("no arguments returns zero value", func(t *testing.T) {
		t.Parallel()

		got := Coalesce[int]()
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	t.Run("all empty returns zero value", func(t *testing.T) {
		t.Parallel()

		got := Coalesce(0, 0, 0)
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	t.Run("returns first non-empty", func(t *testing.T) {
		t.Parallel()

		got := Coalesce(0, 0, 3, 4)
		if got != 3 {
			t.Fatalf("got %d, want 3", got)
		}
	})

	t.Run("with strings", func(t *testing.T) {
		t.Parallel()

		got := Coalesce("", "", "hello")
		if got != "hello" {
			t.Fatalf("got %q, want %q", got, "hello")
		}
	})
}

func TestEqual(t *testing.T) {
	t.Parallel()

	t.Run("equal values", func(t *testing.T) {
		t.Parallel()

		if !Equal("a", "a") {
			t.Fatal("expected true")
		}
	})

	t.Run("different values", func(t *testing.T) {
		t.Parallel()

		if Equal("a", "b") {
			t.Fatal("expected false")
		}
	})

	t.Run("equal slices", func(t *testing.T) {
		t.Parallel()

		if !Equal([]int{1, 2}, []int{1, 2}) {
			t.Fatal("expected true")
		}
	})
}

func TestNotEqual(t *testing.T) {
	t.Parallel()

	t.Run("equal values", func(t *testing.T) {
		t.Parallel()

		if NotEqual("a", "a") {
			t.Fatal("expected false")
		}
	})

	t.Run("different values", func(t *testing.T) {
		t.Parallel()

		if !NotEqual("a", "b") {
			t.Fatal("expected true")
		}
	})
}

func TestNil(t *testing.T) {
	t.Parallel()

	t.Run("untyped nil", func(t *testing.T) {
		t.Parallel()

		if !Nil(nil) {
			t.Fatal("expected true")
		}
	})

	t.Run("typed nil pointer", func(t *testing.T) {
		t.Parallel()

		var p *int
		if !Nil(p) {
			t.Fatal("expected true")
		}
	})

	t.Run("non-nil value", func(t *testing.T) {
		t.Parallel()

		if Nil(10) {
			t.Fatal("expected false")
		}
	})
}

func TestNotNil(t *testing.T) {
	t.Parallel()

	t.Run("nil value", func(t *testing.T) {
		t.Parallel()

		var p *int
		if NotNil(p) {
			t.Fatal("expected false")
		}
	})

	t.Run("non-nil value", func(t *testing.T) {
		t.Parallel()

		if !NotNil(10) {
			t.Fatal("expected true")
		}
	})
}

func TestEmpty(t *testing.T) {
	t.Parallel()

	t.Run("no arguments", func(t *testing.T) {
		t.Parallel()

		if !Empty() {
			t.Fatal("expected true")
		}
	})

	t.Run("single empty string", func(t *testing.T) {
		t.Parallel()

		if !Empty("") {
			t.Fatal("expected true")
		}
	})

	t.Run("single non-empty string", func(t *testing.T) {
		t.Parallel()

		if Empty("x") {
			t.Fatal("expected false")
		}
	})

	t.Run("all empty values", func(t *testing.T) {
		t.Parallel()

		if !Empty("", 0, nil) {
			t.Fatal("expected true")
		}
	})

	t.Run("mixed with non-empty", func(t *testing.T) {
		t.Parallel()

		if Empty("x", "") {
			t.Fatal("expected false")
		}
	})
}

func TestNotEmpty(t *testing.T) {
	t.Parallel()

	t.Run("no arguments", func(t *testing.T) {
		t.Parallel()

		if !NotEmpty() {
			t.Fatal("expected true")
		}
	})

	t.Run("single non-empty value", func(t *testing.T) {
		t.Parallel()

		if !NotEmpty("x") {
			t.Fatal("expected true")
		}
	})

	t.Run("single empty value", func(t *testing.T) {
		t.Parallel()

		if NotEmpty("") {
			t.Fatal("expected false")
		}
	})

	t.Run("all non-empty values", func(t *testing.T) {
		t.Parallel()

		if !NotEmpty("x", "y", "z") {
			t.Fatal("expected true")
		}
	})

	t.Run("mixed with empty", func(t *testing.T) {
		t.Parallel()

		if NotEmpty("x", "", "y") {
			t.Fatal("expected false")
		}
	})
}

// --- String functions ---

func TestRandomString(t *testing.T) {
	t.Parallel()

	t.Run("zero length", func(t *testing.T) {
		t.Parallel()

		got := RandomString(0)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("negative length", func(t *testing.T) {
		t.Parallel()

		got := RandomString(-1)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("single char charset", func(t *testing.T) {
		t.Parallel()

		got := RandomString(5, WithCharset("A"))
		if got != "AAAAA" {
			t.Fatalf("got %q, want %q", got, "AAAAA")
		}
	})

	t.Run("default charset produces correct length", func(t *testing.T) {
		t.Parallel()

		got := RandomString(10)
		if len(got) != 10 {
			t.Fatalf("len = %d, want 10", len(got))
		}
	})
}

func TestSubstring(t *testing.T) {
	t.Parallel()

	t.Run("basic extraction", func(t *testing.T) {
		t.Parallel()

		got := Substring("hello", 0, 2)
		if got != "he" {
			t.Fatalf("got %q, want %q", got, "he")
		}
	})

	t.Run("offset at middle", func(t *testing.T) {
		t.Parallel()

		got := Substring("hello", 2, 3)
		if got != "llo" {
			t.Fatalf("got %q, want %q", got, "llo")
		}
	})

	t.Run("offset beyond length", func(t *testing.T) {
		t.Parallel()

		got := Substring("hello", 10, 2)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("length exceeds remaining", func(t *testing.T) {
		t.Parallel()

		got := Substring("hello", 3, 100)
		if got != "lo" {
			t.Fatalf("got %q, want %q", got, "lo")
		}
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()

		got := Substring("", 0, 5)
		if got != "" {
			t.Fatalf("got %q, want empty", got)
		}
	})

	t.Run("unicode characters", func(t *testing.T) {
		t.Parallel()

		got := Substring("héllo", 0, 3)
		if got != "hél" {
			t.Fatalf("got %q, want %q", got, "hél")
		}
	})

	t.Run("full string", func(t *testing.T) {
		t.Parallel()

		got := Substring("hello", 0, 5)
		if got != "hello" {
			t.Fatalf("got %q, want %q", got, "hello")
		}
	})
}

func TestChunkString(t *testing.T) {
	t.Parallel()

	t.Run("zero size", func(t *testing.T) {
		t.Parallel()

		got := ChunkString("abc", 0)
		if !reflect.DeepEqual(got, []string{""}) {
			t.Fatalf("got %v, want [\"\"]", got)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()

		got := ChunkString("", 3)
		if !reflect.DeepEqual(got, []string{""}) {
			t.Fatalf("got %v, want [\"\"]", got)
		}
	})

	t.Run("size greater than string length", func(t *testing.T) {
		t.Parallel()

		got := ChunkString("abc", 5)
		if !reflect.DeepEqual(got, []string{"abc"}) {
			t.Fatalf("got %v, want [\"abc\"]", got)
		}
	})

	t.Run("even split", func(t *testing.T) {
		t.Parallel()

		got := ChunkString("abcdef", 2)

		want := []string{"ab", "cd", "ef"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("uneven split", func(t *testing.T) {
		t.Parallel()

		got := ChunkString("abcde", 2)

		want := []string{"ab", "cd", "e"}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}

func TestCapitalize(t *testing.T) {
	t.Parallel()

	t.Run("lowercase input", func(t *testing.T) {
		t.Parallel()

		got := Capitalize("hello")
		if got != "Hello" {
			t.Fatalf("got %q, want %q", got, "Hello")
		}
	})

	t.Run("mixed case input", func(t *testing.T) {
		t.Parallel()

		got := Capitalize("hELLO")
		if got != "Hello" {
			t.Fatalf("got %q, want %q", got, "Hello")
		}
	})
}

func TestWords(t *testing.T) {
	t.Parallel()

	t.Run("camelCase split", func(t *testing.T) {
		t.Parallel()

		got := Words("helloWorld")

		want := []string{"hello", "World"}
		if !slices.Equal(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("uppercase with numbers", func(t *testing.T) {
		t.Parallel()

		got := Words("HTTPServer2ID")

		want := []string{"HTTP", "Server", "2", "ID"}
		if !slices.Equal(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("special characters become separators", func(t *testing.T) {
		t.Parallel()

		got := Words("ab-__cd!!ef")

		want := []string{"ab", "cd", "ef"}
		if !slices.Equal(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("mixed with underscores and numbers", func(t *testing.T) {
		t.Parallel()

		got := Words("v1_test")

		want := []string{"v", "1", "test"}
		if !slices.Equal(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("number to letter transition", func(t *testing.T) {
		t.Parallel()

		got := Words("12xy")

		want := []string{"12", "xy"}
		if !slices.Equal(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}

func TestPascalCase(t *testing.T) {
	t.Parallel()

	t.Run("from kebab case", func(t *testing.T) {
		t.Parallel()

		got := PascalCase("hello-world")
		if got != "HelloWorld" {
			t.Fatalf("got %q, want %q", got, "HelloWorld")
		}
	})

	t.Run("from snake case with numbers", func(t *testing.T) {
		t.Parallel()

		got := PascalCase("hello_world_user42")
		if got != "HelloWorldUser42" {
			t.Fatalf("got %q, want %q", got, "HelloWorldUser42")
		}
	})
}

func TestCamelCase(t *testing.T) {
	t.Parallel()

	t.Run("from snake case", func(t *testing.T) {
		t.Parallel()

		got := CamelCase("hello_world")
		if got != "helloWorld" {
			t.Fatalf("got %q, want %q", got, "helloWorld")
		}
	})

	t.Run("from spaces with numbers", func(t *testing.T) {
		t.Parallel()

		got := CamelCase("hello_world user42")
		if got != "helloWorldUser42" {
			t.Fatalf("got %q, want %q", got, "helloWorldUser42")
		}
	})
}

func TestKebabCase(t *testing.T) {
	t.Parallel()

	t.Run("from pascal case", func(t *testing.T) {
		t.Parallel()

		got := KebabCase("HelloWorld")
		if got != "hello-world" {
			t.Fatalf("got %q, want %q", got, "hello-world")
		}
	})

	t.Run("from pascal case with numbers", func(t *testing.T) {
		t.Parallel()

		got := KebabCase("HelloWorld User42")
		if got != "hello-world-user-42" {
			t.Fatalf("got %q, want %q", got, "hello-world-user-42")
		}
	})
}

func TestSnakeCase(t *testing.T) {
	t.Parallel()

	t.Run("from pascal case", func(t *testing.T) {
		t.Parallel()

		got := SnakeCase("HelloWorld")
		if got != "hello_world" {
			t.Fatalf("got %q, want %q", got, "hello_world")
		}
	})

	t.Run("from pascal case with numbers", func(t *testing.T) {
		t.Parallel()

		got := SnakeCase("HelloWorld User42")
		if got != "hello_world_user_42" {
			t.Fatalf("got %q, want %q", got, "hello_world_user_42")
		}
	})
}

// --- Slice functions ---

func TestFilterBy(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		var empty []int

		got := FilterBy(empty, func(x int) bool { return true })
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})

	t.Run("filter even numbers", func(t *testing.T) {
		t.Parallel()

		nums := []int{1, 2, 3, 4, 5}

		got := FilterBy(nums, func(x int) bool { return x%2 == 0 })
		if !slices.Equal(got, []int{2, 4}) {
			t.Fatalf("got %v, want [2 4]", got)
		}
	})

	t.Run("filter keeps all", func(t *testing.T) {
		t.Parallel()

		nums := []int{2, 4, 6}

		got := FilterBy(nums, func(x int) bool { return x%2 == 0 })
		if !slices.Equal(got, []int{2, 4, 6}) {
			t.Fatalf("got %v, want [2 4 6]", got)
		}
	})

	t.Run("filter removes all", func(t *testing.T) {
		t.Parallel()

		nums := []int{1, 3, 5}

		got := FilterBy(nums, func(x int) bool { return x%2 == 0 })
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
}

func TestCount(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := Count([]int{})
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()

		got := Count([]int{1, 2, 3})
		if got != 3 {
			t.Fatalf("got %d, want 3", got)
		}
	})
}

func TestCountBy(t *testing.T) {
	t.Parallel()

	even := func(x int) bool { return x%2 == 0 }

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := CountBy([]int{}, even)
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	t.Run("count even numbers", func(t *testing.T) {
		t.Parallel()

		got := CountBy([]int{1, 2, 3, 4, 5}, even)
		if got != 2 {
			t.Fatalf("got %d, want 2", got)
		}
	})

	t.Run("does not mutate input slice", func(t *testing.T) {
		t.Parallel()

		original := []int{1, 2, 3, 4, 5}
		snapshot := Copy(original)
		CountBy(original, even)

		if !slices.Equal(original, snapshot) {
			t.Fatalf("input was mutated: got %v, want %v", original, snapshot)
		}
	})
}

func TestToMap(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := ToMap([]string{})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty map", got)
		}
	})

	t.Run("unique elements", func(t *testing.T) {
		t.Parallel()

		got := ToMap([]string{"a", "b", "c"})
		if got["a"] != 1 || got["b"] != 1 || got["c"] != 1 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("duplicate elements", func(t *testing.T) {
		t.Parallel()

		got := ToMap([]string{"a", "b", "a"})
		if got["a"] != 2 || got["b"] != 1 {
			t.Fatalf("got %v", got)
		}
	})
}

func TestToMapBy(t *testing.T) {
	t.Parallel()

	even := func(x int) bool { return x%2 == 0 }

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := ToMapBy([]int{}, even)
		if len(got) != 0 {
			t.Fatalf("got %v, want empty map", got)
		}
	})

	t.Run("filter even numbers", func(t *testing.T) {
		t.Parallel()

		got := ToMapBy([]int{1, 2, 3, 4}, even)
		if len(got) != 2 || got[2] != 1 || got[4] != 1 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("no match", func(t *testing.T) {
		t.Parallel()

		got := ToMapBy([]int{1, 3, 5}, even)
		if len(got) != 0 {
			t.Fatalf("got %v, want empty map", got)
		}
	})
}

func TestIn(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		if In(1) {
			t.Fatal("expected false")
		}
	})

	t.Run("value found", func(t *testing.T) {
		t.Parallel()

		if !In(2, 1, 2, 3) {
			t.Fatal("expected true")
		}
	})

	t.Run("value not found", func(t *testing.T) {
		t.Parallel()

		if In(9, 1, 2, 3) {
			t.Fatal("expected false")
		}
	})
}

func TestNotIn(t *testing.T) {
	t.Parallel()

	t.Run("value found", func(t *testing.T) {
		t.Parallel()

		if NotIn(2, 1, 2) {
			t.Fatal("expected false")
		}
	})

	t.Run("value not found", func(t *testing.T) {
		t.Parallel()

		if !NotIn(9, 1, 2) {
			t.Fatal("expected true")
		}
	})
}

func TestEvery(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		if Every([]int{1, 2}) {
			t.Fatal("expected false")
		}
	})

	t.Run("all present", func(t *testing.T) {
		t.Parallel()

		if !Every([]int{1, 2}, 1, 2, 3) {
			t.Fatal("expected true")
		}
	})

	t.Run("some missing", func(t *testing.T) {
		t.Parallel()

		if Every([]int{1, 4}, 1, 2, 3) {
			t.Fatal("expected false")
		}
	})
}

func TestSome(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		if Some([]int{1, 2}) {
			t.Fatal("expected false")
		}
	})

	t.Run("some present", func(t *testing.T) {
		t.Parallel()

		if !Some([]int{4, 2}, 1, 2, 3) {
			t.Fatal("expected true")
		}
	})

	t.Run("none present", func(t *testing.T) {
		t.Parallel()

		if Some([]int{4, 5}, 1, 2, 3) {
			t.Fatal("expected false")
		}
	})
}

func TestNone(t *testing.T) {
	t.Parallel()

	t.Run("none present", func(t *testing.T) {
		t.Parallel()

		if !None([]int{4, 5}, 1, 2, 3) {
			t.Fatal("expected true")
		}
	})

	t.Run("some present", func(t *testing.T) {
		t.Parallel()

		if None([]int{1, 5}, 1, 2, 3) {
			t.Fatal("expected false")
		}
	})
}

func TestUnion(t *testing.T) {
	t.Parallel()

	t.Run("disjoint slices", func(t *testing.T) {
		t.Parallel()

		got := Union([]int{1, 2}, []int{3, 4})
		if !sameMultiset(got, []int{1, 2, 3, 4}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("overlapping slices", func(t *testing.T) {
		t.Parallel()

		got := Union([]int{1, 2}, []int{2, 3})
		if !sameMultiset(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("duplicates in input", func(t *testing.T) {
		t.Parallel()

		got := Union([]int{1, 2, 2}, []int{2, 3})
		if !sameMultiset(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("empty slices", func(t *testing.T) {
		t.Parallel()

		got := Union([]int{}, []int{})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
}

func TestIntersection(t *testing.T) {
	t.Parallel()

	t.Run("overlapping slices", func(t *testing.T) {
		t.Parallel()

		got := Intersection([]int{1, 2, 3}, []int{2, 3, 4})
		if !sameMultiset(got, []int{2, 3}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("disjoint slices", func(t *testing.T) {
		t.Parallel()

		got := Intersection([]int{1, 2}, []int{3, 4})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})

	t.Run("empty slices", func(t *testing.T) {
		t.Parallel()

		got := Intersection([]int{}, []int{1, 2})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
}

func TestDifference(t *testing.T) {
	t.Parallel()

	t.Run("overlapping slices", func(t *testing.T) {
		t.Parallel()

		left, right := Difference([]int{1, 2, 3}, []int{2, 3, 4})
		if !sameMultiset(left, []int{1}) || !sameMultiset(right, []int{4}) {
			t.Fatalf("left=%v, right=%v", left, right)
		}
	})

	t.Run("identical slices", func(t *testing.T) {
		t.Parallel()

		left, right := Difference([]int{1, 2}, []int{1, 2})
		if len(left) != 0 || len(right) != 0 {
			t.Fatalf("left=%v, right=%v", left, right)
		}
	})

	t.Run("disjoint slices", func(t *testing.T) {
		t.Parallel()

		left, right := Difference([]int{1, 2}, []int{3, 4})
		if !sameMultiset(left, []int{1, 2}) || !sameMultiset(right, []int{3, 4}) {
			t.Fatalf("left=%v, right=%v", left, right)
		}
	})
}

func TestMap(t *testing.T) {
	t.Parallel()

	t.Run("double values", func(t *testing.T) {
		t.Parallel()

		double := func(x int) int { return x * 2 }

		got := Map([]int{1, 2, 3}, double)
		if !slices.Equal(got, []int{2, 4, 6}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		identity := func(x int) int { return x }

		got := Map([]int{}, identity)
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
}

func TestMapBy(t *testing.T) {
	t.Parallel()

	double := func(x int) int { return x * 2 }
	keepOdd := func(x int) bool { return x%2 == 1 }

	t.Run("with filter", func(t *testing.T) {
		t.Parallel()

		got := MapBy([]int{1, 2, 3}, double, keepOdd)
		if !slices.Equal(got, []int{2, 0, 6}) {
			t.Fatalf("got %v, want [2 0 6]", got)
		}
	})

	t.Run("all kept", func(t *testing.T) {
		t.Parallel()

		all := func(x int) bool { return true }

		got := MapBy([]int{1, 2, 3}, double, all)
		if !slices.Equal(got, []int{2, 4, 6}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestCopy(t *testing.T) {
	t.Parallel()

	t.Run("nil slice", func(t *testing.T) {
		t.Parallel()

		if Copy[int](nil) != nil {
			t.Fatal("expected nil")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := Copy([]int{})
		if got == nil || len(got) != 0 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("non-empty slice", func(t *testing.T) {
		t.Parallel()

		src := []int{1, 2, 3}

		got := Copy(src)
		if !slices.Equal(got, src) {
			t.Fatalf("got %v, want %v", got, src)
		}
	})

	t.Run("independent from source", func(t *testing.T) {
		t.Parallel()

		src := []int{1, 2, 3}
		got := Copy(src)
		got[0] = 99

		if src[0] != 1 {
			t.Fatal("copy is not independent from source")
		}
	})
}

func TestReverse(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := Reverse([]int{})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		got := Reverse([]int{1})
		if !slices.Equal(got, []int{1}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		t.Parallel()

		got := Reverse(Copy([]int{1, 2, 3}))
		if !slices.Equal(got, []int{3, 2, 1}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestShuffle(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := Shuffle([]int{})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		got := Shuffle([]int{7})
		if !slices.Equal(got, []int{7}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("preserves elements", func(t *testing.T) {
		t.Parallel()

		src := []int{1, 2, 3, 4, 5}

		got := Shuffle(Copy(src))
		if !sameMultiset(got, src) {
			t.Fatalf("shuffle lost elements: got %v", got)
		}
	})
}

func TestDeduplicate(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := Deduplicate([]int{})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		got := Deduplicate([]int{1})
		if !slices.Equal(got, []int{1}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("no duplicates", func(t *testing.T) {
		t.Parallel()

		got := Deduplicate([]int{1, 2, 3})
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("with duplicates", func(t *testing.T) {
		t.Parallel()

		got := Deduplicate([]int{1, 2, 2, 3, 1})
		if !sameMultiset(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestSort(t *testing.T) {
	t.Parallel()

	t.Run("unsorted", func(t *testing.T) {
		t.Parallel()

		got := Sort([]int{3, 1, 2})
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("already sorted", func(t *testing.T) {
		t.Parallel()

		got := Sort([]int{1, 2, 3})
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("reverse sorted", func(t *testing.T) {
		t.Parallel()

		got := Sort([]int{3, 2, 1})
		if !slices.Equal(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestChunk(t *testing.T) {
	t.Parallel()

	t.Run("size zero", func(t *testing.T) {
		t.Parallel()

		got := Chunk([]int{1, 2}, 0)
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})

	t.Run("even split", func(t *testing.T) {
		t.Parallel()

		got := Chunk([]int{1, 2, 3, 4}, 2)

		want := [][]int{{1, 2}, {3, 4}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("uneven split", func(t *testing.T) {
		t.Parallel()

		got := Chunk([]int{1, 2, 3, 4, 5}, 2)

		want := [][]int{{1, 2}, {3, 4}, {5}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("size greater than slice length", func(t *testing.T) {
		t.Parallel()

		got := Chunk([]int{1, 2}, 5)

		want := [][]int{{1, 2}}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		out := Delete(0, []int{})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("negative index", func(t *testing.T) {
		t.Parallel()

		out := Delete(-1, []int{1, 2})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("index out of bounds", func(t *testing.T) {
		t.Parallel()

		out := Delete(5, []int{1, 2})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("valid middle index", func(t *testing.T) {
		t.Parallel()

		got := Delete(1, []int{1, 2, 3})
		if !slices.Equal(got, []int{1, 3}) {
			t.Fatalf("got %v, want [1 3]", got)
		}
	})

	t.Run("first element", func(t *testing.T) {
		t.Parallel()

		got := Delete(0, []int{1, 2, 3})
		if !slices.Equal(got, []int{2, 3}) {
			t.Fatalf("got %v, want [2 3]", got)
		}
	})

	t.Run("last element", func(t *testing.T) {
		t.Parallel()

		got := Delete(2, []int{1, 2, 3})
		if !slices.Equal(got, []int{1, 2}) {
			t.Fatalf("got %v, want [1 2]", got)
		}
	})
}

func TestDeleteRange(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		out := DeleteRange(0, 1, []int{})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("negative start", func(t *testing.T) {
		t.Parallel()

		out := DeleteRange(-1, 1, []int{1, 2})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("start out of bounds", func(t *testing.T) {
		t.Parallel()

		out := DeleteRange(5, 6, []int{1, 2})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("negative end", func(t *testing.T) {
		t.Parallel()

		out := DeleteRange(0, -1, []int{1, 2})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("end out of bounds", func(t *testing.T) {
		t.Parallel()

		out := DeleteRange(0, 5, []int{1, 2})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("start greater than end", func(t *testing.T) {
		t.Parallel()

		out := DeleteRange(2, 1, []int{1, 2, 3})
		if out != nil {
			t.Fatalf("got %v, want nil", out)
		}
	})

	t.Run("valid range", func(t *testing.T) {
		t.Parallel()

		got := DeleteRange(1, 2, []int{1, 2, 3, 4})
		if !slices.Equal(got, []int{1, 4}) {
			t.Fatalf("got %v, want [1 4]", got)
		}
	})
}

func TestPush(t *testing.T) {
	t.Parallel()

	t.Run("to non-empty slice", func(t *testing.T) {
		t.Parallel()

		got := Push([]int{1}, 2)
		if !slices.Equal(got, []int{1, 2}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("to empty slice", func(t *testing.T) {
		t.Parallel()

		got := Push([]int{}, 1)
		if !slices.Equal(got, []int{1}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestPop(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		v, rest := Pop([]int{})
		if v != 0 || rest != nil {
			t.Fatalf("got v=%d, rest=%v", v, rest)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		v, rest := Pop([]int{7})
		if v != 7 || len(rest) != 0 {
			t.Fatalf("got v=%d, rest=%v", v, rest)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		t.Parallel()

		v, rest := Pop([]int{7, 8})
		if v != 8 || !slices.Equal(rest, []int{7}) {
			t.Fatalf("got v=%d, rest=%v", v, rest)
		}
	})
}

func TestMax(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := Max([]int{})
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		got := Max([]int{5})
		if got != 5 {
			t.Fatalf("got %d, want 5", got)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		t.Parallel()

		got := Max([]int{2, 9, 3})
		if got != 9 {
			t.Fatalf("got %d, want 9", got)
		}
	})
}

func TestMin(t *testing.T) {
	t.Parallel()

	t.Run("empty slice", func(t *testing.T) {
		t.Parallel()

		got := Min([]int{})
		if got != 0 {
			t.Fatalf("got %d, want 0", got)
		}
	})

	t.Run("single element", func(t *testing.T) {
		t.Parallel()

		got := Min([]int{5})
		if got != 5 {
			t.Fatalf("got %d, want 5", got)
		}
	})

	t.Run("multiple elements", func(t *testing.T) {
		t.Parallel()

		got := Min([]int{2, 9, 1, 3})
		if got != 1 {
			t.Fatalf("got %d, want 1", got)
		}
	})
}

// --- Map functions ---

func TestHasKey(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "b": 2}

	t.Run("key exists", func(t *testing.T) {
		t.Parallel()

		if !HasKey("a", m) {
			t.Fatal("expected true")
		}
	})

	t.Run("key missing", func(t *testing.T) {
		t.Parallel()

		if HasKey("z", m) {
			t.Fatal("expected false")
		}
	})
}

func TestKeys(t *testing.T) {
	t.Parallel()

	t.Run("single map", func(t *testing.T) {
		t.Parallel()

		got := Keys(map[string]int{"a": 1, "b": 2})
		if !sameMultiset(got, []string{"a", "b"}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("multiple maps", func(t *testing.T) {
		t.Parallel()

		m1 := map[string]int{"a": 1}
		m2 := map[string]int{"b": 2, "c": 3}

		got := Keys(m1, m2)
		if !sameMultiset(got, []string{"a", "b", "c"}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("overlapping keys", func(t *testing.T) {
		t.Parallel()

		m1 := map[string]int{"a": 1}
		m2 := map[string]int{"a": 2}

		got := Keys(m1, m2)
		if !sameMultiset(got, []string{"a", "a"}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestUniqueKeys(t *testing.T) {
	t.Parallel()

	t.Run("no duplicates", func(t *testing.T) {
		t.Parallel()

		m1 := map[string]int{"a": 1}
		m2 := map[string]int{"b": 2}

		got := UniqueKeys(m1, m2)
		if !sameMultiset(got, []string{"a", "b"}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("with duplicates", func(t *testing.T) {
		t.Parallel()

		m1 := map[string]int{"a": 1, "b": 2}
		m2 := map[string]int{"b": 3, "c": 4}

		got := UniqueKeys(m1, m2)
		if !sameMultiset(got, []string{"a", "b", "c"}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestValues(t *testing.T) {
	t.Parallel()

	t.Run("single map", func(t *testing.T) {
		t.Parallel()

		got := Values(map[string]int{"a": 1, "b": 2})
		if !sameMultiset(got, []int{1, 2}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("multiple maps", func(t *testing.T) {
		t.Parallel()

		m1 := map[string]int{"a": 1}
		m2 := map[string]int{"b": 2, "c": 3}

		got := Values(m1, m2)
		if !sameMultiset(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestUniqueValues(t *testing.T) {
	t.Parallel()

	t.Run("no duplicates", func(t *testing.T) {
		t.Parallel()

		m1 := map[string]int{"a": 1}
		m2 := map[string]int{"b": 2}

		got := UniqueValues(m1, m2)
		if !sameMultiset(got, []int{1, 2}) {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("with duplicates", func(t *testing.T) {
		t.Parallel()

		m1 := map[string]int{"a": 1, "b": 2}
		m2 := map[string]int{"c": 2, "d": 3}

		got := UniqueValues(m1, m2)
		if !sameMultiset(got, []int{1, 2, 3}) {
			t.Fatalf("got %v", got)
		}
	})
}

func TestPickByKeys(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "b": 2, "c": 3}

	t.Run("some keys exist", func(t *testing.T) {
		t.Parallel()

		got := PickByKeys(m, []string{"a", "z"})
		if len(got) != 1 || got["a"] != 1 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("all keys exist", func(t *testing.T) {
		t.Parallel()

		got := PickByKeys(m, []string{"a", "b"})
		if len(got) != 2 || got["a"] != 1 || got["b"] != 2 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("no keys exist", func(t *testing.T) {
		t.Parallel()

		got := PickByKeys(m, []string{"x", "y"})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
}

func TestPickByValues(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "b": 2, "c": 3}

	t.Run("some values match", func(t *testing.T) {
		t.Parallel()

		got := PickByValues(m, []int{1, 9})
		if len(got) != 1 || got["a"] != 1 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("no values match", func(t *testing.T) {
		t.Parallel()

		got := PickByValues(m, []int{8, 9})
		if len(got) != 0 {
			t.Fatalf("got %v, want empty", got)
		}
	})
}

func TestOmitByKeys(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "b": 2, "c": 3}

	t.Run("omit some keys", func(t *testing.T) {
		t.Parallel()

		got := OmitByKeys(m, []string{"b"})
		if len(got) != 2 || got["a"] != 1 || got["c"] != 3 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("omit no keys", func(t *testing.T) {
		t.Parallel()

		got := OmitByKeys(m, []string{"z"})
		if len(got) != 3 {
			t.Fatalf("got %v", got)
		}
	})
}

func TestOmitByValues(t *testing.T) {
	t.Parallel()

	m := map[string]int{"a": 1, "b": 2, "c": 3}

	t.Run("omit some values", func(t *testing.T) {
		t.Parallel()

		got := OmitByValues(m, []int{2})
		if len(got) != 2 || got["a"] != 1 || got["c"] != 3 {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("omit no values", func(t *testing.T) {
		t.Parallel()

		got := OmitByValues(m, []int{9})
		if len(got) != 3 {
			t.Fatalf("got %v", got)
		}
	})
}
