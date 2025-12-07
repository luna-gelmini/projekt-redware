import { marked } from 'marked';
import createDOMPurify from 'dompurify';

// Configure marked
marked.setOptions({
    breaks: true, // Convert \n to <br>
    gfm: true,    // GitHub Flavored Markdown
    headerIds: false,
    mangle: false
});

// Configure DOMPurify
const DOMPurify = createDOMPurify(window);

// Custom renderer to open links in new tab
const renderer = new marked.Renderer();
const linkRenderer = renderer.link;
renderer.link = (href, title, text) => {
    const html = linkRenderer.call(renderer, href, title, text);
    return html.replace(/^<a /, '<a target="_blank" rel="noopener noreferrer" ');
};

marked.use({ renderer });

export function renderMarkdown(text) {
    if (!text) return '';
    const rawHtml = marked.parse(text);
    return DOMPurify.sanitize(rawHtml);
}
