package dedupe

import (
	"context"
	"log"

	"github.com/clive2000/rclone/cmd"
	"github.com/clive2000/rclone/fs/config/flags"
	"github.com/clive2000/rclone/fs/operations"
	"github.com/spf13/cobra"
)

var (
	dedupeMode = operations.DeduplicateInteractive
)

func init() {
	cmd.Root.AddCommand(commandDefinition)
	cmdFlag := commandDefinition.Flags()
	flags.FVarP(cmdFlag, &dedupeMode, "dedupe-mode", "", "Dedupe mode interactive|skip|first|newest|oldest|largest|smallest|rename.")
}

var commandDefinition = &cobra.Command{
	Use:   "dedupe [mode] remote:path",
	Short: `Interactively find duplicate filenames and delete/rename them.`,
	Long: `

By default ` + "`dedupe`" + ` interactively finds files with duplicate
names and offers to delete all but one or rename them to be
different.

This is only useful with backends like Google Drive which can have
duplicate file names. It can be run on wrapping backends (eg crypt) if
they wrap a backend which supports duplicate file names.

In the first pass it will merge directories with the same name.  It
will do this iteratively until all the identically named directories
have been merged.

In the second pass, for every group of duplicate file names, it will
delete all but one identical files it finds without confirmation.
This means that for most duplicated files the ` + "`dedupe`" + `
command will not be interactive.

` + "`dedupe`" + ` considers files to be identical if they have the
same file path and the same hash. If the backend does not support hashes (eg crypt wrapping
Google Drive) then they will never be found to be identical. If you
use the ` + "`--size-only`" + ` flag then files will be considered
identical if they have the same size (any hash will be ignored). This
can be useful on crypt backends which do not support hashes.

**Important**: Since this can cause data loss, test first with the
` + "`--dry-run` or the `--interactive`/`-i`" + ` flag.

Here is an example run.

Before - with duplicates

    $ rclone lsl drive:dupes
      6048320 2016-03-05 16:23:16.798000000 one.txt
      6048320 2016-03-05 16:23:11.775000000 one.txt
       564374 2016-03-05 16:23:06.731000000 one.txt
      6048320 2016-03-05 16:18:26.092000000 one.txt
      6048320 2016-03-05 16:22:46.185000000 two.txt
      1744073 2016-03-05 16:22:38.104000000 two.txt
       564374 2016-03-05 16:22:52.118000000 two.txt

Now the ` + "`dedupe`" + ` session

    $ rclone dedupe drive:dupes
    2016/03/05 16:24:37 Google drive root 'dupes': Looking for duplicates using interactive mode.
    one.txt: Found 4 files with duplicate names
    one.txt: Deleting 2/3 identical duplicates (MD5 "1eedaa9fe86fd4b8632e2ac549403b36")
    one.txt: 2 duplicates remain
      1:      6048320 bytes, 2016-03-05 16:23:16.798000000, MD5 1eedaa9fe86fd4b8632e2ac549403b36
      2:       564374 bytes, 2016-03-05 16:23:06.731000000, MD5 7594e7dc9fc28f727c42ee3e0749de81
    s) Skip and do nothing
    k) Keep just one (choose which in next step)
    r) Rename all to be different (by changing file.jpg to file-1.jpg)
    s/k/r> k
    Enter the number of the file to keep> 1
    one.txt: Deleted 1 extra copies
    two.txt: Found 3 files with duplicates names
    two.txt: 3 duplicates remain
      1:       564374 bytes, 2016-03-05 16:22:52.118000000, MD5 7594e7dc9fc28f727c42ee3e0749de81
      2:      6048320 bytes, 2016-03-05 16:22:46.185000000, MD5 1eedaa9fe86fd4b8632e2ac549403b36
      3:      1744073 bytes, 2016-03-05 16:22:38.104000000, MD5 851957f7fb6f0bc4ce76be966d336802
    s) Skip and do nothing
    k) Keep just one (choose which in next step)
    r) Rename all to be different (by changing file.jpg to file-1.jpg)
    s/k/r> r
    two-1.txt: renamed from: two.txt
    two-2.txt: renamed from: two.txt
    two-3.txt: renamed from: two.txt

The result being

    $ rclone lsl drive:dupes
      6048320 2016-03-05 16:23:16.798000000 one.txt
       564374 2016-03-05 16:22:52.118000000 two-1.txt
      6048320 2016-03-05 16:22:46.185000000 two-2.txt
      1744073 2016-03-05 16:22:38.104000000 two-3.txt

Dedupe can be run non interactively using the ` + "`" + `--dedupe-mode` + "`" + ` flag or by using an extra parameter with the same value

  * ` + "`" + `--dedupe-mode interactive` + "`" + ` - interactive as above.
  * ` + "`" + `--dedupe-mode skip` + "`" + ` - removes identical files then skips anything left.
  * ` + "`" + `--dedupe-mode first` + "`" + ` - removes identical files then keeps the first one.
  * ` + "`" + `--dedupe-mode newest` + "`" + ` - removes identical files then keeps the newest one.
  * ` + "`" + `--dedupe-mode oldest` + "`" + ` - removes identical files then keeps the oldest one.
  * ` + "`" + `--dedupe-mode largest` + "`" + ` - removes identical files then keeps the largest one.
  * ` + "`" + `--dedupe-mode smallest` + "`" + ` - removes identical files then keeps the smallest one.
  * ` + "`" + `--dedupe-mode rename` + "`" + ` - removes identical files then renames the rest to be different.

For example to rename all the identically named photos in your Google Photos directory, do

    rclone dedupe --dedupe-mode rename "drive:Google Photos"

Or

    rclone dedupe rename "drive:Google Photos"
`,
	Run: func(command *cobra.Command, args []string) {
		cmd.CheckArgs(1, 2, command, args)
		if len(args) > 1 {
			err := dedupeMode.Set(args[0])
			if err != nil {
				log.Fatal(err)
			}
			args = args[1:]
		}
		fdst := cmd.NewFsSrc(args)
		cmd.Run(false, false, command, func() error {
			return operations.Deduplicate(context.Background(), fdst, dedupeMode)
		})
	},
}
