package fs

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FS struct {
	Dir        string
	Root       *Node
	CurrentDir string
	Current    *Node
}

func NewFS(dir string) *FS {
	return &FS{
		Dir:        dir,
		CurrentDir: dir,
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

var (
	ErrNotADirectory = fmt.Errorf("not a directory")
	ErrDoesntExist   = fmt.Errorf("does not exist")
	ErrNotFound      = fmt.Errorf("not found")
)

func (fs *FS) ChangeDir(dir string) error {
	var newStrDir string
	if dir == ".." {
		if fs.CurrentDir == fs.Dir {
			// already at root
			return nil
		}

		newStrDir = filepath.Dir(fs.CurrentDir)
		if newStrDir == fs.Dir {
			// return to root
			fs.CurrentDir = fs.Dir
			fs.Current = fs.Root
			return nil
		}

	} else {
		newStrDir = filepath.Join(fs.CurrentDir, dir)
	}

	searchDir := strings.TrimPrefix(newStrDir, fs.Dir)
	searchDir = strings.TrimPrefix(searchDir, string(filepath.Separator))

	newDir := fs.Root.Find(strings.Split(searchDir, string(filepath.Separator)))
	if newDir == nil {
		return ErrNotFound
	}
	if !newDir.IsDir {
		return ErrNotADirectory
	}

	fs.CurrentDir = newStrDir
	fs.Current = newDir
	return nil
}

type ListItem struct {
	Name  string
	Size  uint64
	IsDir bool
}

type ListView struct {
	Items     []ListItem
	TotalSize uint64
}

func (fs *FS) List() *ListView {
	list := &ListView{}
	if fs.CurrentDir != fs.Dir {
		list.Items = append(list.Items, ListItem{
			Name:  "..",
			Size:  0,
			IsDir: true,
		})
	}

	for _, node := range fs.Current.Child {
		list.Items = append(list.Items, ListItem{
			Name:  node.Name,
			Size:  node.Size,
			IsDir: node.IsDir,
		})
		list.TotalSize += node.Size
	}

	// sort list
	sort.Sort(list)

	return list
}

// Implement ListView Sort interface
func (l *ListView) Len() int {
	return len(l.Items)
}

// Direcories sorted by size desc, then files sorted by size desc
// ORDER BY is_dir DESC, size DESC
// ".." is always first
func (l *ListView) Less(i, j int) bool {
	if l.Items[i].Name == ".." {
		return true
	}
	if l.Items[j].Name == ".." {
		return false
	}

	if l.Items[i].IsDir && !l.Items[j].IsDir {
		return true
	}
	if !l.Items[i].IsDir && l.Items[j].IsDir {
		return false
	}

	if l.Items[i].Size > l.Items[j].Size {
		return true
	}
	if l.Items[i].Size < l.Items[j].Size {
		return false
	}

	return l.Items[i].Name < l.Items[j].Name
}

func (l *ListView) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

// Node represents a file or directory in the filesystem
type Node struct {
	Name  string
	IsDir bool
	Size  uint64
	Files uint64
	Child map[string]*Node
}

// add adds a file to the tree
// path is the path to the file, split into components: ["a", "b", "c"]
// info is the file info for the file at the end of the path
// node size is the sum of the sizes of all files in the subtree rooted at the node
// leaf node is a file, non-leaf node is a directory
func (n *Node) add(path []string, info fs.FileInfo) {
	if len(path) == 0 {
		n.Size = uint64(info.Size())
		n.IsDir = false
		return
	}
	n.Size += uint64(info.Size())
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
