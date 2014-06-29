package turbo

import (
	"strings"
)

type PathTree struct {
	refs map[string]*PathTreeNode
}

type PathTreeNode struct {
	parent   *PathTreeNode
	children map[*PathTreeNode]bool
	path     string
	depth    int
}

func (tree *PathTree) put(path string) {
	node := PathTreeNode{
		parent:   nil,
		children: make(map[*PathTreeNode]bool),
		path:     path,
		depth:    strings.Count(path, "/"),
	}
	tree.refs[path] = &node

	// Check relationship with the heads
	parent := tree.parent(path)
	if parent != nil {
		node.parent = parent
		parent.children[&node] = true
	}
}

func (tree *PathTree) parent(path string) *PathTreeNode {
	depth := strings.Count(path, "/")

	if depth <= 1 {
		return nil
	}

	for i := range depth {
		index := strings.LastIndex(path, "/")
		if index > 0 {
			path = path[:index]
			if tree.refs[path] != nil {
				return tree.refs[path]
			}
		} else {
			return nil
		}
	}

	return nil
}

func (tree *PathTree) children(path string) *map[*PathTreeNode]bool {
	node := tree.refs[path]

	if node == nil {
		return nil
	} else if len(node.children) == 0 {
		return nil
	} else {
		return &node.children
	}
}

func (tree *PathTree) allChildren(path string) *map[string]bool {
	node := tree.refs[path]
	list := make(map[string]bool)
	if node != nil {
		tree.allChildrenHelper(list, node)
		return &list
	} else {
		return nil
	}
}

func (tree *PathTree) kill(path string) {
	node := tree.refs[path]
	list := make(map[string]bool)
	if node != nil {
		tree.killHelper(list, node)
		return &list
	} else {
		return nil
	}
}

func (tree *PathTree) allChildrenHelper(list *map[string]bool, node *PathTreeNode) {
	for child, _ := range node.children {
		allChildrenHelper(list, child)
		list[child.path] = true
	}
}
