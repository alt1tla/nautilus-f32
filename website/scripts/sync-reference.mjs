// Copies canonical docs out of the repo's docs/ folder into the site's
// content collection, prepending the frontmatter Starlight requires.
// docs/functions.md stays the single source of truth — the copy under
// src/content/docs/ is generated and gitignored.
import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const root = dirname(dirname(fileURLToPath(import.meta.url)));

const pages = [
  {
    src: join(root, '..', 'docs', 'functions.md'),
    dest: join(root, 'src', 'content', 'docs', 'reference', 'functions.md'),
    description:
      'How a scan evaluates each IEC 61131-3 language, and every built-in operator, function, and function block.',
  },
];

for (const page of pages) {
  const raw = readFileSync(page.src, 'utf8');
  const lines = raw.split('\n');
  const title = lines[0].replace(/^#\s*/, '').trim();
  const body = lines.slice(1).join('\n').trimStart();
  const out = `---\ntitle: ${JSON.stringify(title)}\ndescription: ${JSON.stringify(page.description)}\n---\n\n${body}`;
  mkdirSync(dirname(page.dest), { recursive: true });
  writeFileSync(page.dest, out);
  console.log(`synced ${page.src} -> ${page.dest}`);
}
