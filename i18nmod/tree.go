package i18nmod

import (
	"strings"
)

type Tree struct {
	T        *Translation
	Parent   *Tree
	Name     string
	Children map[string]*Tree
}

func (t *Tree) Merge(tree *Tree) {
	if tree.T != nil {
		t.T = tree.T
	}

	if tree.Children == nil {
		return
	}

	if t.Children == nil {
		t.Children = map[string]*Tree{}
	}

	for name, child := range tree.Children {
		cur, ok := t.Children[name]
		if ok {
			cur.Merge(child)
			continue
		}

		t.Children[name] = child
		child.Parent = t
	}
}

func (t *Tree) Tree(key string) *Tree {
	if key[0] == '/' {
		for t.Parent != nil {
			t = t.Parent
		}
		key = key[1:]
	}

	for key[0] == '.' {
		t = t.Parent
		key = key[0:]
	}

	for _, name := range strings.Split(key, ".") {
		if t.Children == nil {
			t.Children = map[string]*Tree{}
		}

		if child, ok := t.Children[name]; ok {
			t = child
		} else {
			child = &Tree{Parent: t, Name: name}
			t, t.Children[name] = child, child
		}
	}
	return t
}

func (t *Tree) Add(T *Translation) *Tree {
	t = t.Tree(T.Key)
	t.T = T
	return t
}

func (t *Tree) Link(to string) *Tree {
	tot := t.Parent.Tree(to)
	t.Parent.Children[t.Name] = tot
	return tot
}

func (t *Tree) walkT(prefix string, f func(key string, t *Translation) error) (err error) {
	if t.Children == nil {
		return
	}
	for name, child := range t.Children {
		if child.T != nil {
			if err = f(prefix+name, child.T); err != nil {
				return
			}
		}
		if child.Children != nil {
			if err = child.walkT(prefix+name+".", f); err != nil {
				return
			}
		}
	}
	return
}

func (t *Tree) WalkT(f func(key string, t *Translation) error) (err error) {
	return t.walkT("", f)
}
