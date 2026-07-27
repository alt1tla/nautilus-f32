import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://nautilus.joyautomation.com',
  integrations: [
    starlight({
      title: 'nautilus',
      description:
        'SCADA as software — a Go + SvelteKit toolkit for building industrial control and supervisory systems with version control, tests, CI/CD, and code review.',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/joyautomation/nautilus',
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/joyautomation/nautilus/edit/main/website/',
      },
      sidebar: [
        {
          label: 'Start here',
          items: [{ label: 'Getting started', slug: 'getting-started' }],
        },
        {
          label: 'Guides',
          autogenerate: { directory: 'guides' },
        },
        {
          label: 'Reference',
          autogenerate: { directory: 'reference' },
        },
      ],
    }),
  ],
});
