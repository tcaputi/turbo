package turbo

import (
	"strings"
)

type PathTree struct {
	refs map[string]*PathTreeNode
}

func NewPathTree() *PathTree {
	tree := PathTree{
		refs: make(map[string]*PathTreeNode),
	}
	return &tree
}

type PathTreeNode struct {
	tree     *PathTree
	parent   *PathTreeNode
	children map[*PathTreeNode]bool
	path     string
	depth    int
}

func (tree *PathTree) path(path string) string {
	return "/" + strings.Trim(path, "/")
}

func (tree *PathTree) get(path string) *PathTreeNode {
	if tree.refs[path] == nil {
		return nil
	} else {
		return tree.refs[path]
	}
}

func (tree *PathTree) put(path string) *PathTreeNode {
	path = tree.path(path)
	// First check if this path represents a node already
	if tree.refs[path] != nil {
		return tree.refs[path]
	}

	node := PathTreeNode{
		tree:     tree,
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

func (node *PathTreeNode) remove() {
	// Append children to parent
	for nodeChild, _ := range node.children {
		node.parent.children[nodeChild] = true
		nodeChild.parent = node.parent
		delete(node.children, nodeChild)
	}
	node.destroy()
}

// Deletes this node from the tree - does not consider children
func (node *PathTreeNode) destroy() {
	// Orphan this node
	delete(node.parent.children, node)
	node.parent = nil
	// Remove from refs
	delete(node.tree.refs, node.path)
}

func (node *PathTreeNode) parentIsRoot() bool {
	return node.parent.path == "/"
}

func (node *PathTreeNode) cascade(iterator func(*PathTreeNode)) {
	for child, _ := range node.children {
		child.cascade(iterator)
	}
	iterator(node)
}

func (node *PathTreeNode) hasImmediateParent() bool {
	return strings.LastIndex(node.path, SLASH) == (len(node.parent.path) - 1)
}
