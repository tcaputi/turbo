package turbo

import (
	"testing"
)

var (
	tree = &PathTree{
		refs: make(map[string]*PathTreeNode),
	}
)

func TestPut(t *testing.T) {
	tree.put("a/b/c")
	tree.put("a/b/c/d/e/f/g/h")
	tree.put("a/b/c/d/e/f")
	tree.put("a/b/c/d/e/f/g/h")
	tree.put("a/b/c/d/e/f/g/h/i/j")
	tree.put("a/b/c/d")
}

func TestParent(t *testing.T) {
	if tree.parent("a/b/c/d").path != "a/b/c" {
		t.Fail()
	}
	if tree.parent("a/b/c/d/e/f/g/h").path != "a/b/c/d/e/f" {
		t.Fail()
	}
	if tree.parent("a/b/c/d/e/f/g/h/i/j").path != "a/b/c/d/e/f/g/h" {
		t.Fail()
	}
	if tree.parent("a") != nil {
		t.Fail()
	}
}

func TestChildren(t *testing.T) {
}
