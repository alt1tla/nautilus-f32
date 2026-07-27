<script lang="ts">
	// Tag bindings: prop → tag rows. Commits the whole map on each completed
	// change (the doc's bind object is small); an in-progress new row lives
	// locally until both halves are filled. Tag inputs complete from the
	// live controller snapshot (datalist#mimic-tags); "!" prefix negates.
	let {
		binds,
		hints = [],
		onchange
	}: {
		binds: Record<string, string> | undefined;
		/** Suggested bindable prop names for this component. */
		hints?: string[];
		onchange: (next: Record<string, string> | null) => void;
	} = $props();

	const rows = $derived(Object.entries(binds ?? {}));
	let draftProp = $state('');
	let draftTag = $state('');

	function commit(entries: [string, string][]) {
		const next = Object.fromEntries(entries.filter(([p, t]) => p !== '' && t !== ''));
		onchange(Object.keys(next).length ? next : null);
	}

	function setRow(i: number, prop: string, tag: string) {
		const next = rows.map((r, j) => (j === i ? ([prop, tag] as [string, string]) : r));
		commit(next);
	}

	function removeRow(i: number) {
		commit(rows.filter((_, j) => j !== i));
	}

	function tryAddDraft() {
		if (draftProp === '' || draftTag === '') return;
		commit([...rows, [draftProp, draftTag]]);
		draftProp = '';
		draftTag = '';
	}
</script>

<div class="binds">
	<span class="head">bind</span>
	{#each rows as [prop, tag], i (i)}
		<div class="row">
			<input
				class="nx-input"
				value={prop}
				list="mimic-props"
				aria-label="Bound prop"
				onchange={(e) => setRow(i, e.currentTarget.value, tag)}
			/>
			<span class="arrow">←</span>
			<input
				class="nx-input"
				value={tag}
				list="mimic-tags"
				aria-label="Tag name"
				onchange={(e) => setRow(i, prop, e.currentTarget.value)}
			/>
			<button class="x" title="Remove binding" onclick={() => removeRow(i)}>×</button>
		</div>
	{/each}
	<div class="row">
		<input
			class="nx-input"
			placeholder="prop"
			list="mimic-props"
			aria-label="New bound prop"
			bind:value={draftProp}
			onchange={tryAddDraft}
		/>
		<span class="arrow">←</span>
		<input
			class="nx-input"
			placeholder="tag"
			list="mimic-tags"
			aria-label="New tag name"
			bind:value={draftTag}
			onchange={tryAddDraft}
		/>
		<span class="x pad"></span>
	</div>
	<datalist id="mimic-props">
		{#each hints as h (h)}<option value={h}></option>{/each}
	</datalist>
</div>

<style>
	.binds {
		display: grid;
		gap: 4px;
	}
	.head {
		font-size: 11px;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--nx-muted);
	}
	.row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr) 18px;
		gap: 4px;
		align-items: center;
	}
	.row :global(.nx-input) {
		width: 100%;
		min-width: 0;
	}
	.arrow {
		color: var(--nx-muted);
		font-size: 11px;
	}
	.x {
		border: none;
		background: transparent;
		color: var(--nx-muted);
		cursor: pointer;
		font-size: 13px;
		padding: 0;
		line-height: 1;
	}
	.x:hover {
		color: var(--nx-err);
	}
	.x.pad {
		cursor: default;
	}
</style>
