<script lang="ts">
	// Instruction palette: pick a template, fill its fields here, Insert posts
	// an insertStatement op — Go validates the fragment before it touches the
	// file, and focus never leaves the diagram. Name-like fields complete
	// against declared tags (and function/type vocabularies) — free text with
	// suggestions, never a gate.
	import { postOp, type FbdEditOp } from './vscodeApi';
	import type { VarDecl } from './layout';
	import Popover from './Popover.svelte';
	import Suggest from './Suggest.svelte';
	import { FUNCTIONS, FB_TYPES, TYPES, type SuggestItem } from './suggest';

	let { open = $bindable(false), vars = [] }: { open?: boolean; vars?: VarDecl[] } = $props();

	// What a field completes against: declared tags (optionally comma-lists),
	// callable functions, FB types, or declarable types.
	type Kind = 'tags' | 'tags-multi' | 'fn' | 'fbtype' | 'type';
	type Field = { key: string; def: string; kind?: Kind };
	type Template = {
		label: string;
		preview: string;
		fields: Field[];
		// Either a netlist statement (insertStatement) or a custom op.
		build?: (f: Record<string, string>) => string;
		op?: (f: Record<string, string>) => FbdEditOp;
	};
	const TEMPLATES: Template[] = [
		{
			label: 'block → wire',
			preview: 'w = AND(a, b)',
			fields: [
				{ key: 'name', def: 'w1' },
				{ key: 'function', def: 'AND', kind: 'fn' },
				{ key: 'inputs', def: 'in1, in2', kind: 'tags-multi' }
			],
			build: (f) => `${f.name} = ${f.function}(${f.inputs})`
		},
		{
			label: 'coil (assign output)',
			preview: 'Out := src',
			fields: [
				{ key: 'output', def: 'Output', kind: 'tags' },
				{ key: 'source', def: 'source', kind: 'tags' }
			],
			build: (f) => `${f.output} := ${f.source}`
		},
		{
			label: 'timer',
			preview: 't : TON(…)',
			fields: [
				{ key: 'name', def: 't1' },
				{ key: 'type', def: 'TON', kind: 'fbtype' },
				{ key: 'IN', def: 'condition', kind: 'tags' },
				{ key: 'PT', def: 'T#1S' }
			],
			build: (f) => `${f.name} : ${f.type}(IN := ${f.IN}, PT := ${f.PT})`
		},
		{
			label: 'counter',
			preview: 'c : CTU(…)',
			fields: [
				{ key: 'name', def: 'c1' },
				{ key: 'CU', def: 'count', kind: 'tags' },
				{ key: 'R', def: 'reset', kind: 'tags' },
				{ key: 'PV', def: '10' }
			],
			build: (f) => `${f.name} : CTU(CU := ${f.CU}, R := ${f.R}, PV := ${f.PV})`
		},
		{
			label: 'comment',
			preview: '// note',
			fields: [{ key: 'text', def: 'note' }],
			build: (f) => '// ' + f.text
		},
		{
			label: 'input reference (bare)',
			preview: 'chip: name →',
			fields: [{ key: 'name', def: 'Tag1', kind: 'tags' }],
			// A ghost layout entry: the chip exists on the canvas only until a
			// wire makes it real netlist text.
			op: (f) => ({ type: 'setLayout', entries: [{ node: 'g:in.' + f.name, x: 40, y: 40 }] })
		},
		{
			label: 'output reference (bare)',
			preview: '→ coil: name',
			fields: [{ key: 'name', def: 'Out1', kind: 'tags' }],
			op: (f) => ({ type: 'setLayout', entries: [{ node: 'g:out.' + f.name, x: 240, y: 40 }] })
		},
		{
			label: 'variable (external tag)',
			preview: 'name : REAL',
			fields: [
				{ key: 'name', def: 'Tag1' },
				{ key: 'type', def: 'REAL', kind: 'type' }
			],
			op: (f) => ({ type: 'declareVar', newName: f.name, value: f.type, text: 'VAR_EXTERNAL' })
		},
		{
			label: 'local variable (retained)',
			preview: 'VAR name : REAL',
			fields: [
				{ key: 'name', def: 'local1' },
				{ key: 'type', def: 'REAL', kind: 'type' }
			],
			op: (f) => ({ type: 'declareVar', newName: f.name, value: f.type, text: 'VAR' })
		}
	];

	const tagItems = $derived<SuggestItem[]>(vars.map((v) => ({ name: v.name, detail: v.type })));
	function itemsFor(kind: Kind): SuggestItem[] {
		switch (kind) {
			case 'tags':
			case 'tags-multi':
				return tagItems;
			case 'fn':
				return FUNCTIONS;
			case 'fbtype':
				return FB_TYPES;
			case 'type':
				return TYPES;
		}
	}

	let active = $state<Template | null>(null);
	let values = $state<Record<string, string>>({});

	function pick(t: Template) {
		active = t;
		values = Object.fromEntries(t.fields.map((f) => [f.key, f.def]));
	}
	function commit() {
		if (!active) return;
		postOp(active.op ? active.op(values) : { type: 'insertStatement', text: active.build!(values) });
		open = false;
		active = null;
	}
	function keydown(ev: KeyboardEvent) {
		ev.stopPropagation();
		if (ev.key === 'Enter') commit();
		if (ev.key === 'Escape') {
			if (active) active = null;
			else open = false;
		}
	}
</script>

{#if open}
	<Popover onkeydown={keydown}>
		{#if !active}
			{#each TEMPLATES as t (t.label)}
				<button class="item" onclick={() => pick(t)}>
					<span>{t.label}</span>
					<code>{t.preview}</code>
				</button>
			{/each}
		{:else}
			{#each active.fields as f (f.key)}
				<label class="field">
					<span>{f.key}</span>
					{#if f.kind}
						<Suggest
							cls="grow"
							bind:value={values[f.key]}
							items={itemsFor(f.kind)}
							multi={f.kind === 'tags-multi'}
						/>
					{:else}
						<input class="nx-input grow" spellcheck="false" bind:value={values[f.key]} />
					{/if}
				</label>
			{/each}
			<div class="actions">
				<button onclick={() => (active = null)}>back</button>
				<button class="primary" onclick={commit}>insert</button>
			</div>
		{/if}
	</Popover>
{/if}

<style>
	.item {
		display: flex;
		justify-content: space-between;
		gap: 12px;
		background: transparent;
		border: none;
		color: var(--nx-ui-ink);
		padding: 5px 8px;
		border-radius: 3px;
		cursor: pointer;
		font-size: 12px;
		text-align: left;
	}
	.item:hover {
		background: var(--nx-hover);
	}
	.item code {
		font-family: var(--nx-mono);
		font-size: 10px;
		color: var(--nx-muted);
	}
	.field {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 3px 8px;
		font-size: 11px;
		color: var(--nx-muted);
	}
	.field span {
		min-width: 64px;
	}
	.field :global(.grow) {
		flex: 1;
	}
	.actions {
		display: flex;
		justify-content: flex-end;
		gap: 6px;
		padding: 6px 8px 4px;
	}
	.actions button {
		padding: 2px 10px;
		border-radius: 3px;
		border: 1px solid var(--nx-border);
		background: transparent;
		color: var(--nx-ui-ink);
		cursor: pointer;
		font-size: 12px;
	}
	.actions button.primary {
		background: var(--nx-btn-bg);
		color: var(--nx-btn-ink);
		border-color: transparent;
	}
</style>
