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

func (tree *PathTree) path(path string) string {
	return "/" + strings.Trim(path, "/")
}

func (tree *PathTree) put(path string) *PathTreeNode {
	path = tree.path(path)
	// First check if this path represents a node already
	if tree.refs[path] != nil {
		return tree.refs[path]
	}

	node := PathTreeNode{
		parent:   nil,
		children: make(map[*PathTreeNode]bool),
		path:     path,
		depth:    strings.Count(path, "/"),
	}
	tree.refs[path] = &node
	// Check relationship with the heads
	parent := tree.parent(path)
	if parent == nil {
		parent = &PathTreeNode{
			parent:   nil,
			children: make(map[*PathTreeNode]bool),
			path:     "/",
			depth:    0,
		}
		tree.refs["/"] = parent
	}

	node.parent = parent
	// Check if new parent has any of the new node's children
	for child, _ := range parent.children {
		if child != nil {
			if child.depth >= (node.depth + 1) {
				// Potential child of new node
				if strings.Index(child.path, node.path) == 0 {
					// Child should have this node as a parent
					delete(parent.children, child)
					child.parent = &node
					node.children[child] = true
				}
			}
		}
	}
	parent.children[&node] = true

	return &node
}

func (tree *PathTree) parent(path string) *PathTreeNode {
	path = tree.path(path)
	// First check if this path represents a node already
	node := tree.refs[path]
	if node != nil && node.parent != nil {
		return node.parent
	}
	// Check if its the root path
	if path == "/" {
		return nil
	}
	// Establish depth to loop
	depth := strings.Count(path, "/")
	// Recursively check every parent path
	for i := 0; i < depth; i++ {
		index := strings.LastIndex(path, "/")
		if index != -1 {
			path = path[:index]
			// Mitigate against surface level paths
			if len(path) == 0 {
				path = "/"
			}
			// Return the parent if it exists
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

func (tree *PathTree) kill(path string) *map[string]bool {
	node := tree.refs[path]
	list := make(map[string]bool)
	if node != nil {
		// TODO
		// tree.killHelper(list, node)
		return &list
	} else {
		return nil
	}
}
