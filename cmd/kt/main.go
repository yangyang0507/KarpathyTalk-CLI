package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/yangyang0507/KarpathyTalk-CLI/internal/client"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/display"
	"github.com/yangyang0507/KarpathyTalk-CLI/internal/tui"
	"golang.org/x/term"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func terminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// boolFlagger is satisfied by flag.BoolVar values, which consume no next argument.
type boolFlagger interface{ IsBoolFlag() bool }

// splitArgs separates args into positional arguments and parses flags into fs.
// It correctly keeps flag-value pairs together, supporting any argument order:
// both "kt post 5 --markdown" and "kt post --markdown 5" work as expected.
func splitArgs(fs *flag.FlagSet, args []string) (positional []string) {
	var flagArgs []string
	i := 0
	for i < len(args) {
		a := args[i]
		if len(a) == 0 || a[0] != '-' {
			positional = append(positional, a)
			i++
			continue
		}
		flagArgs = append(flagArgs, a)
		name := strings.TrimLeft(a, "-")
		if strings.ContainsRune(name, '=') {
			// --flag=value form: value already embedded, no look-ahead needed.
			i++
			continue
		}
		// Look up whether this flag accepts a value argument.
		if f := fs.Lookup(name); f != nil {
			bf, ok := f.Value.(boolFlagger)
			isBool := ok && bf.IsBoolFlag()
			if !isBool && i+1 < len(args) && len(args[i+1]) > 0 && args[i+1][0] != '-' {
				flagArgs = append(flagArgs, args[i+1])
				i += 2
				continue
			}
		}
		i++
	}
	fs.Parse(flagArgs) //nolint:errcheck // ExitOnError mode exits on parse failure
	return
}

func main() {
	// Global --host flag (must appear before subcommand)
	host := flag.String("host", client.DefaultHost, "API root URL (e.g. http://localhost:8080)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Usage = usage
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}
	client.SetVersion(version)

	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	c := client.New(*host)
	cmd, rest := args[0], args[1:]

	switch cmd {
	case "timeline":
		runTimeline(c, rest)
	case "user":
		runUser(c, rest)
	case "post":
		runPost(c, rest)
	case "docs":
		runDocs(c)
	case "help":
		runHelp(rest)
	default:
		fmt.Fprintf(os.Stderr, "kt: unknown command %q\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `kt — KarpathyTalk CLI (%s)

Usage:
  kt [--host <url>] <command> [flags]

Commands:
  timeline            Browse the public timeline
  user <username>     View a user's profile and posts
  post <id>           View a single post and its replies
  docs                Fetch platform API documentation

Global Flags:
  --host <url>   API root URL (default: https://karpathytalk.com)
  --version      Print version and exit

Run "kt help <command>" for command-specific flags and examples.
`, version)
}

func usageTimeline() {
	fmt.Fprint(os.Stderr, `Browse the public timeline (root posts only).

Usage:
  kt timeline [flags]

Flags:
  --limit <n>      Posts per page (default: 20, max: 100)
  --before <id>    Pagination cursor: load posts older than this ID
  --json           Output raw JSON
  --markdown       Output post content as Markdown

Examples:
  kt timeline
  kt timeline --limit 50
  kt timeline --before 230
  kt timeline --json | jq '.posts[].author.username'
  kt timeline --markdown > posts.md
`)
}

func usageUser() {
	fmt.Fprint(os.Stderr, `View a user's profile and their posts.

Usage:
  kt user <username> [flags]

Flags:
  --replies        Show only replies (default: root posts only)
  --limit <n>      Posts per page (default: 20)
  --before <id>    Pagination cursor
  --json           Output raw JSON
  --markdown       Output user profile Markdown with YAML frontmatter

Examples:
  kt user karpathy
  kt user karpathy --replies
  kt user karpathy --json | jq '.posts | length'
  kt user karpathy --markdown > karpathy.md
`)
}

func usagePost() {
	fmt.Fprint(os.Stderr, `View a single post and its direct replies.

Usage:
  kt post <id> [flags]

Flags:
  --limit <n>       Replies per page (default: 20)
  --json            Output raw JSON
  --markdown        Post Markdown with YAML frontmatter
  --raw             Raw post Markdown, no frontmatter
  --revision <n>    View a specific revision (use with --markdown or --raw)

Examples:
  kt post 42
  kt post 42 --raw | llm "summarize this"
  kt post 42 --markdown > post-42.md
  kt post 42 --revision 2 --raw
`)
}

func usageDocs() {
	fmt.Fprint(os.Stderr, `Fetch the KarpathyTalk API documentation as Markdown.

Usage:
  kt docs

Examples:
  kt docs
  kt docs | llm "what API endpoints are available?"
`)
}

func runHelp(args []string) {
	if len(args) == 0 {
		usage()
		return
	}
	switch args[0] {
	case "timeline":
		usageTimeline()
	case "user":
		usageUser()
	case "post":
		usagePost()
	case "docs":
		usageDocs()
	default:
		fmt.Fprintf(os.Stderr, "kt: unknown command %q\n\n", args[0])
		usage()
		os.Exit(1)
	}
}

func runTimeline(c *client.Client, args []string) {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	fs.Usage = usageTimeline
	limit := fs.Int("limit", 20, "Posts per page (max 100)")
	before := fs.Int64("before", 0, "Pagination cursor: load posts older than this ID")
	asJSON := fs.Bool("json", false, "Output raw JSON")
	asMarkdown := fs.Bool("markdown", false, "Output post content as Markdown")
	splitArgs(fs, args)

	// TUI path: skip fetch, hand all params to the TUI which loads its own data.
	if !*asJSON && !*asMarkdown && isTTY() {
		if err := tui.Run(tui.Config{
			Mode:   "timeline",
			Limit:  *limit,
			Before: *before,
			Client: c,
		}); err != nil {
			fatal(err)
		}
		return
	}

	hasParent := false
	resp, err := c.GetPosts(client.PostsQuery{
		Limit:     *limit,
		Before:    *before,
		HasParent: &hasParent,
	})
	if err != nil {
		fatal(err)
	}

	switch {
	case *asJSON:
		display.PrintJSON(resp)
	case *asMarkdown:
		for _, p := range resp.Posts {
			display.PrintMarkdown(p.ContentMarkdown)
			fmt.Println("---")
		}
	default:
		for _, p := range resp.Posts {
			display.PostSummary(p)
		}
		if resp.HasMore {
			display.Pagination(resp.NextCursor)
		}
	}
}

func runUser(c *client.Client, args []string) {
	fs := flag.NewFlagSet("user", flag.ExitOnError)
	fs.Usage = usageUser
	replies := fs.Bool("replies", false, "Show only replies (default: root posts only)")
	limit := fs.Int("limit", 20, "Posts per page")
	before := fs.Int64("before", 0, "Pagination cursor")
	asJSON := fs.Bool("json", false, "Output raw JSON")
	asMarkdown := fs.Bool("markdown", false, "Output user profile Markdown with YAML frontmatter")
	positional := splitArgs(fs, args)

	if len(positional) < 1 {
		usageUser()
		os.Exit(1)
	}
	username := positional[0]

	// Markdown uses a dedicated server endpoint, handle before TTY check.
	if *asMarkdown {
		md, err := c.GetUserMarkdown(username)
		if err != nil {
			fatal(err)
		}
		display.PrintMarkdown(md)
		return
	}

	// TUI path: skip fetch, hand all params to the TUI which loads its own data.
	if !*asJSON && isTTY() {
		if err := tui.Run(tui.Config{
			Mode:     "user",
			Username: username,
			Limit:    *limit,
			Before:   *before,
			Replies:  *replies,
			Client:   c,
		}); err != nil {
			fatal(err)
		}
		return
	}

	user, err := c.GetUser(username)
	if err != nil {
		fatal(err)
	}

	hasParent := *replies
	resp, err := c.GetPosts(client.PostsQuery{
		Author:    username,
		Limit:     *limit,
		Before:    *before,
		HasParent: &hasParent,
	})
	if err != nil {
		fatal(err)
	}

	if *asJSON {
		display.PrintJSON(map[string]any{"user": user, "posts": resp.Posts, "has_more": resp.HasMore, "next_cursor": resp.NextCursor})
		return
	}

	display.UserProfile(*user)
	for _, p := range resp.Posts {
		display.PostSummary(p)
	}
	if resp.HasMore {
		display.Pagination(resp.NextCursor)
	}
}

func runPost(c *client.Client, args []string) {
	fs := flag.NewFlagSet("post", flag.ExitOnError)
	fs.Usage = usagePost
	limit := fs.Int("limit", 20, "Replies per page")
	asJSON := fs.Bool("json", false, "Output raw JSON")
	asMarkdown := fs.Bool("markdown", false, "Post Markdown with YAML frontmatter")
	asRaw := fs.Bool("raw", false, "Raw post Markdown, no frontmatter")
	revision := fs.Int("revision", 0, "View a specific revision (use with --markdown or --raw)")
	positional := splitArgs(fs, args)

	if len(positional) < 1 {
		usagePost()
		os.Exit(1)
	}
	id, err := strconv.ParseInt(positional[0], 10, 64)
	if err != nil || id < 1 {
		fmt.Fprintln(os.Stderr, "kt post: invalid post ID")
		usagePost()
		os.Exit(1)
	}

	if *revision > 0 && !*asMarkdown && !*asRaw {
		fmt.Fprintln(os.Stderr, "kt post: --revision only applies with --markdown or --raw")
		os.Exit(1)
	}

	if *asMarkdown {
		md, err := c.GetPostMarkdown(id, *revision)
		if err != nil {
			fatal(err)
		}
		display.PrintMarkdown(md)
		return
	}
	if *asRaw {
		raw, err := c.GetPostRaw(id, *revision)
		if err != nil {
			fatal(err)
		}
		display.PrintMarkdown(raw)
		return
	}

	// TUI path: skip fetch, hand all params to the TUI which loads its own data.
	if !*asJSON && isTTY() {
		if err := tui.Run(tui.Config{
			Mode:       "post",
			PostID:     id,
			ReplyLimit: *limit,
			Client:     c,
		}); err != nil {
			fatal(err)
		}
		return
	}

	post, err := c.GetPost(id)
	if err != nil {
		fatal(err)
	}

	parentID := post.ID
	hasParent := true
	repliesResp, err := c.GetPosts(client.PostsQuery{
		ParentPostID: &parentID,
		HasParent:    &hasParent,
		Limit:        *limit,
	})
	if err != nil {
		fatal(err)
	}

	if *asJSON {
		display.PrintJSON(postJSONOutput(post, repliesResp))
		return
	}

	display.PostFull(*post)
	if len(repliesResp.Posts) > 0 {
		fmt.Printf("── %d direct repl", len(repliesResp.Posts))
		if len(repliesResp.Posts) == 1 {
			fmt.Println("y ──")
		} else {
			fmt.Println("ies ──")
		}
		if repliesResp.HasMore {
			fmt.Printf("   (showing %d of more — use --limit to fetch more)\n", len(repliesResp.Posts))
		}
		fmt.Println()
		for _, r := range repliesResp.Posts {
			display.PostSummary(r)
		}
	}
}

func postJSONOutput(post *client.Post, repliesResp *client.PostsResponse) map[string]any {
	return map[string]any{
		"post":        post,
		"replies":     repliesResp.Posts,
		"has_more":    repliesResp.HasMore,
		"next_cursor": repliesResp.NextCursor,
	}
}

func runDocs(c *client.Client) {
	docs, err := c.GetDocs()
	if err != nil {
		fatal(err)
	}
	if isTTY() {
		fmt.Print(display.RenderMarkdownWidth(docs, terminalWidth()))
		return
	}
	display.PrintMarkdown(docs)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "kt:", err)
	os.Exit(1)
}
