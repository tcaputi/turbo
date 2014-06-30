package turbo

import (
	"testing"
)

func TestParent(t *testing.T) {
	tree := &PathTree{
		refs: make(map[string]*PathTreeNode),
	}

	tree.put("/a/b/c")
	tree.put("a/b/c/d/e/f/g/h/")
	tree.put("/a/b/c/d/e/f")
	tree.put("/a/b/c/d/e/f/g/h")
	tree.put("a/b/c/d/e/f/1")
	tree.put("a/b/c/d/e/f/1/2")
	tree.put("a/b/c/d/e/f/g/h/i/j")
	tree.put("/a/b/c/d/")

	par1 := tree.parent("a/b/c/d")
	par2 := tree.parent("a/b/c/d/e/f/g/h/")
	par3 := tree.parent("/a/b/c/d/e/f/g/h/i/j")
	par4 := tree.parent("a")

	if par1 == nil {
		t.Error("par1 was nil")
		t.FailNow()
	}
	if par2 == nil {
		t.Error("par2 was nil")
		t.FailNow()
	}
	if par3 == nil {
		t.Error("par3 was nil")
		t.FailNow()
	}
	if par4 == nil {
		t.Error("par4 was nil")
		t.FailNow()
	}

	if par1.path != "/a/b/c" {
		t.Error("par1 path was wrong", par1.path)
	}
	if par2.path != "/a/b/c/d/e/f" {
		t.Error("par2 path was wrong", par2.path)
	}
	if par3.path != "/a/b/c/d/e/f/g/h" {
		t.Error("par3 path was wrong", par3.path)
	}
	if par4.path != "/" {
		t.Error("par4 path was wrong", par4.path)
	}
}

func TestChildren(t *testing.T) {
	tree := &PathTree{
		refs: make(map[string]*PathTreeNode),
	}

	tree.put("/a/b/c")
	tree.put("a/b/c/d/e/f/g/h/")
	tree.put("/a/b/c/d/e/f")
	tree.put("/a/b/c/d/e/f/g/h")
	tree.put("a/b/c/d/e/f/1")
	tree.put("a/b/c/d/e/f/1/2")
	tree.put("a/b/c/d/e/f/g/h/i/j")
	tree.put("/a/b/c/d/")

	chil1 := tree.children("/a/b/c/d")
	chil2 := tree.children("/a/b/c/d/e/f")
	chil3 := tree.children("/a/b/c/d/e/f/g/h")
	chil4 := tree.children("/a/b")

	if chil1 == nil {
		t.Error("chil1 was nil", tree.refs)
		t.FailNow()
	}
	if chil2 == nil {
		t.Error("chil2 was nil")
		t.FailNow()
	}
	if chil3 == nil {
		t.Error("chil3 was nil")
		t.FailNow()
	}

	if len(*chil1) != 1 {
		t.Error("chil1", len(*chil1))
	}
	if len(*chil2) != 2 {
		t.Error("chil2", len(*chil2))
	}
	if len(*chil3) != 1 {
		t.Error("chil3", len(*chil3))
	}
	if chil4 != nil {
		t.Error("chil4", chil4)
	}
}
