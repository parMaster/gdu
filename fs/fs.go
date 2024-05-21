package fs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type FS struct {
	Dir     string
	Root    *Node
	Current *Node
}

func NewFS(dir string) *FS {
	return &FS{
		Dir: dir,
	}
}

func (fs *FS) Scan() error {

	if _, err := os.Stat(fs.Dir); os.IsNotExist(err) {
		return fmt.Errorf("directory %s does not exist", fs.Dir)
	}

	fs.Root = &Node{Name: fs.Dir, IsDir: true}

	filepath.Walk(fs.Dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		path = strings.TrimPrefix(path, fs.Dir)
		path = strings.TrimPrefix(path, string(filepath.Separator))

		if !info.IsDir() {
			fs.Root.add(strings.Split(path, string(filepath.Separator)), info)
		}

		return nil
	})

	fs.Current = fs.Root

	return nil
}

func (n *Node) Find(paths []string) *Node {
	if len(paths) == 0 {
		return n
	}

	child, ok := n.Child[paths[0]]
	if !ok {
		return nil
	}

	return child.Find(paths[1:])
}

type Node struct {
	Name  string
	IsDir bool
	Size  int64
	Files int64
	Child map[string]*Node
}

// add adds a file to the tree
// path is the path to the file, split into components: ["a", "b", "c"]
// info is the file info for the file at the end of the path
// node size is the sum of the sizes of all files in the subtree rooted at the node
// leaf node is a file, non-leaf node is a directory
func (n *Node) add(path []string, info fs.FileInfo) {
	if len(path) == 0 {
		n.Size = info.Size()
		n.IsDir = false
		return
	}
	n.Size += info.Size()
	n.Files++

	name := path[0]
	if n.Child == nil {
		n.Child = map[string]*Node{}
	}

	child, ok := n.Child[name]
	if !ok {
		child = &Node{Name: name, IsDir: true, Size: 0, Files: 0}
		n.Child[name] = child
	}

	child.add(path[1:], info)
}
