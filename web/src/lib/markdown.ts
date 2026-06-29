import { marked } from 'marked'

// Renders markdown to HTML for the document viewer. Content is the user's own
// (single-user, local tool), so it is rendered directly.
export function renderMarkdown(md: string): string {
  return marked.parse(md ?? '', { async: false, breaks: true }) as string
}
