import { mount } from 'svelte';
import '../theme.css';
import './mimic.css';
import MimicApp from './MimicApp.svelte';
import ComponentApp from './ComponentApp.svelte';

// One bundle, two editors: the extension host stamps which document type
// opened it onto <body data-mimic-mode> (mimicEditor.ts / componentEditor.ts)
// — "mimic" mounts the full P&ID editor, "component" mounts the standalone
// *.component.json editor. They're different enough in shape (many
// equipment + pipes vs. one component's dots) to stay separate Svelte
// trees, but small enough, and share enough (registry.ts, portsGestures.ts,
// theme), that a second webview bundle/vite config would be pure overhead.
const mode = document.body.dataset.mimicMode === 'component' ? 'component' : 'mimic';
mount(mode === 'component' ? ComponentApp : MimicApp, { target: document.getElementById('app')! });
