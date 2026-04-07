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
  timeline   Browse the public timeline
  user       View a user's profile and posts
  post       View a single post and its replies
  docs       Fetch platform API documentation

Global Flags:
  --host <url>   API root URL (default: https://karpathytalk.com)
  --version      Print version and exit
`, version)
}

func runTimeline(c *client.Client, args []string) {
	fs := flag.NewFlagSet("timeline", flag.ExitOnError)
	limit := fs.Int("limit", 20, "Number of posts per page (max 100)")
	before := fs.Int64("before", 0, "Pagination cursor: load posts before this ID")
	asJSON := fs.Bool("json", false, "Output raw JSON")
	asMarkdown := fs.Bool("markdown", false, "Output Markdown")
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
	replies := fs.Bool("replies", false, "Show only replies (default: root posts)")
	limit := fs.Int("limit", 20, "Number of posts per page")
	before := fs.Int64("before", 0, "Pagination cursor")
	asJSON := fs.Bool("json", false, "Output raw JSON")
	asMarkdown := fs.Bool("markdown", false, "Output user profile Markdown")
	positional, flags := splitArgs(args)
	fs.Parse(flags)

	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "usage: kt user <username> [flags]")
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
	limit := fs.Int("limit", 20, "Replies per page")
	asJSON := fs.Bool("json", false, "Output raw JSON")
	asMarkdown := fs.Bool("markdown", false, "Output post Markdown with frontmatter")
	asRaw := fs.Bool("raw", false, "Output raw post Markdown (no frontmatter)")
	revision := fs.Int("revision", 0, "View specific revision (use with --markdown/--raw)")
	positional, flags := splitArgs(args)
	fs.Parse(flags)

	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "usage: kt post <id> [flags]")
		os.Exit(1)
	}
	id, err := strconv.ParseInt(positional[0], 10, 64)
	if err != nil || id < 1 {
		fmt.Fprintln(os.Stderr, "kt post: invalid post ID")
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
