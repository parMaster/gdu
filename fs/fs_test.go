package fs

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Fill_Scan_Find(t *testing.T) {

	fs := NewFS("/tmp/nonexistent")
	err := fs.Scan()
	assert.Error(t, err)

	fs = NewFS("/tmp/test")

	os.MkdirAll("/tmp/test/dir1", 0755)
	os.MkdirAll("/tmp/test/dir2/dir2_1", 0755)
	os.MkdirAll("/tmp/test/dir2/dir2_2", 0755)
	os.MkdirAll("/tmp/test/dir3", 0755)

	os.WriteFile("/tmp/test/file1", []byte("file1"), 0644)
	os.WriteFile("/tmp/test/file2", []byte("file2"), 0644)
	os.WriteFile("/tmp/test/dir1/file1", []byte("dir1/file1"), 0644)
	os.WriteFile("/tmp/test/dir1/file2", []byte("dir1/file2"), 0644)
	os.WriteFile("/tmp/test/dir2/dir2_1/file1", []byte("dir2/dir2_1/file1"), 0644)
	os.WriteFile("/tmp/test/dir2/dir2_1/file2", []byte("dir2/dir2_1/file2"), 0644)
	os.WriteFile("/tmp/test/dir2/dir2_2/file1", []byte("dir2/dir2_2/file1"), 0644)
	os.WriteFile("/tmp/test/dir2/dir2_2/file2", []byte("dir2/dir2_2/file2"), 0644)
	os.WriteFile("/tmp/test/dir3/file1", []byte("dir3/file1"), 0644)
	os.WriteFile("/tmp/test/dir3/file2", []byte("dir3/file2"), 0644)

	fs.Scan()

	fs.Root.printTree("\t")

	assert.Equal(t, fs.Current, fs.Root)

	assert.Equal(t, "/tmp/test", fs.Root.Name)

	assert.Equal(t, 5, len(fs.Root.Child))
	assert.Equal(t, 2, len(fs.Root.Child["dir1"].Child))
	assert.Equal(t, 2, len(fs.Root.Child["dir2"].Child))
	assert.Equal(t, 2, len(fs.Root.Child["dir2"].Child["dir2_1"].Child))

	fs.Current = fs.Root.Find([]string{"dir2", "dir2_1"})
	assert.Equal(t, fs.Root.Child["dir2"].Child["dir2_1"], fs.Current)
	assert.Equal(t, "dir2_1", fs.Current.Name)
	assert.Equal(t, 2, len(fs.Current.Child))

}

// helper function that prints the tree
func (n *Node) printTree(indent string) {
	fmt.Printf("%s%s %d (%d files)\n", indent, n.Name, n.Size, n.Files)
	for _, child := range n.Child {
		child.printTree(indent + "  ")
	}
}

// func (n *Node) printLevel(indent string) {
// 	fmt.Printf("%s%s %d\n", indent, n.Name, n.Size)
// 	for _, child := range n.Child {
// 		fmt.Printf("%s%s %d %v (%d files)\n", indent+indent, child.Name, child.Size, child.IsDir, child.Files)
// 	}
// }
