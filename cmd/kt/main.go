package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"golang.org/x/term"
	"kt/internal/client"
	"kt/internal/display"
	"kt/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=<tag>".
var version = "dev"

func isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// splitArgs separates positional arguments from flag arguments.
// This allows "kt post 5 --markdown" and "kt post --markdown 5" to both work.
func splitArgs(args []string) (positional []string, flags []string) {
	for _, a := range args {
		if len(a) > 0 && a[0] == '-' {
			flags = append(flags, a)
		} else {
			positional = append(positional, a)
		}
	}
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
	_, flags := splitArgs(args)
	fs.Parse(flags)

	q := client.PostsQuery{
		Limit:  *limit,
		Before: *before,
	}
	hasParent := false
	q.HasParent = &hasParent

	resp, err := c.GetPosts(q)
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
		if isTTY() {
			if err := tui.Run(tui.Config{Mode: "timeline", Limit: *limit, Client: c}); err != nil {
				fatal(err)
			}
			return
		}
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
	positional, flags := splitArgs(args)
	fs.Parse(flags)

	if len(positional) < 1 {
		usageUser()
		os.Exit(1)
	}
	username := positional[0]

	if *asMarkdown {
		md, err := c.GetUserMarkdown(username)
		if err != nil {
			fatal(err)
		}
		display.PrintMarkdown(md)
		return
	}

	user, err := c.GetUser(username)
	if err != nil {
		fatal(err)
	}

	q := client.PostsQuery{
		Author: username,
		Limit:  *limit,
		Before: *before,
	}
	hasParent := *replies
	q.HasParent = &hasParent

	resp, err := c.GetPosts(q)
	if err != nil {
		fatal(err)
	}

	if *asJSON {
		display.PrintJSON(map[string]any{"user": user, "posts": resp.Posts, "has_more": resp.HasMore, "next_cursor": resp.NextCursor})
		return
	}

	if isTTY() {
		if err := tui.Run(tui.Config{Mode: "user", Username: username, Limit: *limit, Client: c}); err != nil {
			fatal(err)
		}
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
	positional, flags := splitArgs(args)
	fs.Parse(flags)

	if len(positional) < 1 {
		usagePost()
		os.Exit(1)
	}
	id, err := strconv.ParseInt(positional[0], 10, 64)
	if err != nil || id < 1 {
		fmt.Fprintln(os.Stderr, "kt post: invalid post ID\n")
		usagePost()
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

	post, err := c.GetPost(id)
	if err != nil {
		fatal(err)
	}

	parentID := post.ID
	repliesQ := client.PostsQuery{
		ParentPostID: &parentID,
		Limit:        *limit,
	}
	hasParent := true
	repliesQ.HasParent = &hasParent

	repliesResp, err := c.GetPosts(repliesQ)
	if err != nil {
		fatal(err)
	}

	if *asJSON {
		display.PrintJSON(map[string]any{"post": post, "replies": repliesResp.Posts})
		return
	}

	if isTTY() {
		if err := tui.Run(tui.Config{Mode: "post", PostID: id, Client: c}); err != nil {
			fatal(err)
		}
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
		fmt.Println()
		for _, r := range repliesResp.Posts {
			display.PostSummary(r)
		}
	}
}

func runDocs(c *client.Client) {
	docs, err := c.GetDocs()
	if err != nil {
		fatal(err)
	}
	display.PrintMarkdown(docs)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "kt:", err)
	os.Exit(1)
}
