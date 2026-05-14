package commands

import (
	"bytes"
	"fmt"
	"strings"
	"text/tabwriter"
)

func newTabWriter(buf *bytes.Buffer) *tabwriter.Writer {
	return tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
}

// flushTable flushes the tabwriter, then prints the output with the header
// line in bold.
func flushTable(w *tabwriter.Writer, buf *bytes.Buffer) {
	w.Flush()
	output := buf.String()
	idx := strings.IndexByte(output, '\n')
	if idx < 0 {
		fmt.Print(output)
		return
	}
	fmt.Printf("\033[1m%s\033[0m\n", output[:idx])
	fmt.Print(output[idx+1:])
}
