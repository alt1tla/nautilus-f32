<script lang="ts">
	// Showcase for the app-shell primitives: AppShell + MenuBar + Nav for
	// chrome, buttons that open a Modal and fire Toasts, and a gallery of
	// every icon in the kit's curated set.
	import { page } from '$app/state';
	import {
		AppShell,
		MenuBar,
		Nav,
		Modal,
		Toast,
		Tooltip,
		DropdownMenu,
		Button,
		Icon,
		icons,
		theme,
		toast,
		type MenuBarMenu,
		type NavSection
	} from '@joyautomation/nautilus-hmi';

	const menus: MenuBarMenu[] = [
		{
			label: 'File',
			items: [
				{ label: 'New program', icon: 'plus', shortcut: 'Ctrl+N', onSelect: () => toast.info('New program') },
				{ label: 'Open…', icon: 'document', shortcut: 'Ctrl+O', onSelect: () => toast.info('Open…') },
				{ separator: true, label: '' },
				{ label: 'Save', icon: 'clipboard', shortcut: 'Ctrl+S', onSelect: () => toast.success('Saved') },
				{ label: 'Delete', icon: 'trash', disabled: true }
			]
		},
		{
			label: 'View',
			items: [
				{ label: 'Scan diagnostics', icon: 'chart-bar', href: '/' },
				{ label: 'Icon gallery', icon: 'table-cells', onSelect: () => document.getElementById('gallery')?.scrollIntoView({ behavior: 'smooth' }) }
			]
		},
		{
			label: 'Help',
			items: [{ label: 'About nautilus', icon: 'question-mark-circle', onSelect: () => toast.info('nautilus HMI kit — primitives demo') }]
		}
	];

	const navSections: NavSection[] = [
		{
			label: 'Operate',
			items: [
				{ label: 'Heated tank', href: '/', icon: 'fire' },
				{ label: 'Primitives', href: '/primitives', icon: 'cog-6-tooth' },
				{ label: 'Trends', href: '/trends', icon: 'presentation-chart-line' }
			]
		},
		{
			label: 'Monitor',
			items: [
				{ label: 'Alarms', href: '/?tab=alarms', icon: 'bell-alert', badge: 3 },
				{ label: 'Field drivers', href: '/?tab=drivers', icon: 'server' }
			]
		}
	];

	let navCollapsed = $state(false);
	let showModal = $state(false);
	let showDangerModal = $state(false);

	function fireToasts() {
		toast.success('Setpoint written: TempSP = 65.0');
		setTimeout(() => toast.warning('High temperature approaching alarm'), 400);
		setTimeout(() => toast.error('Field driver P101 disconnected'), 900);
		setTimeout(() => toast.info('Reconnecting…'), 1400);
	}

	const iconNames = Object.keys(icons).sort();
</script>

<svelte:head>
	<title>nautilus HMI · primitives</title>
</svelte:head>

<AppShell>
	{#snippet menubar()}
		<MenuBar {menus}>
			{#snippet leading()}
				<strong class="brand">nautilus</strong>
			{/snippet}
		</MenuBar>
	{/snippet}
	{#snippet nav()}
		<Nav sections={navSections} current={page.url.pathname} bind:collapsed={navCollapsed}>
			{#snippet brand()}
				{#if !navCollapsed}<span class="navbrand">HMI primitives</span>{:else}<Icon name="bolt" />{/if}
			{/snippet}
		</Nav>
	{/snippet}

	<div class="demo">
		<header>
			<div>
				<h1>App-shell primitives</h1>
				<p class="sub">Icon · MenuBar · Nav · Modal · Toast · Tooltip · DropdownMenu · Button · AppShell</p>
			</div>
			<div class="chrome">
				<Tooltip text="Toggle theme">
					<Button variant="ghost" icon={theme.effective === 'dark' ? 'sun' : 'moon'} onclick={() => theme.set(theme.effective === 'dark' ? 'light' : 'dark')} aria-label="Toggle theme" />
				</Tooltip>
				<DropdownMenu
					label="Row actions"
					icon="ellipsis-vertical"
					items={[
						{ label: 'Refresh', icon: 'arrow-path', onSelect: () => toast.info('Refreshed') },
						{ label: 'Export', icon: 'document', onSelect: () => toast.success('Exported') },
						{ separator: true, label: '' },
						{ label: 'Remove', icon: 'trash', onSelect: () => showDangerModal = true }
					]}
				/>
			</div>
		</header>

		<section class="card">
			<h2>Buttons &amp; feedback</h2>
			<div class="row">
				<Button variant="primary" icon="check" onclick={() => (showModal = true)}>Open modal</Button>
				<Button variant="secondary" onclick={fireToasts}>Fire toast queue</Button>
				<Button variant="ghost" icon="bell">Ghost</Button>
				<Button variant="danger" icon="trash" onclick={() => (showDangerModal = true)}>Danger</Button>
				<Button variant="secondary" icon="arrow-path" loading>Loading</Button>
				<Button variant="secondary" icon="pencil" disabled>Disabled</Button>
			</div>
		</section>

		<section class="card" id="gallery">
			<h2>Icon gallery ({iconNames.length})</h2>
			<div class="gallery">
				{#each iconNames as name (name)}
					<Tooltip text={name} position="top">
						<div class="icon-tile">
							<Icon {name} size={22} />
						</div>
					</Tooltip>
				{/each}
			</div>
		</section>
	</div>
</AppShell>

<Modal bind:open={showModal} title="Setpoint change" size="sm" onclose={() => {}}>
	<p>Write the new temperature setpoint to the controller?</p>
	<label class="field">
		<span>Temp setpoint (°C)</span>
		<input type="number" value="65" step="0.5" />
	</label>
	{#snippet footer()}
		<Button variant="ghost" onclick={() => (showModal = false)}>Cancel</Button>
		<Button
			variant="primary"
			onclick={() => {
				showModal = false;
				toast.success('Setpoint written');
			}}
		>
			Write
		</Button>
	{/snippet}
</Modal>

<Modal bind:open={showDangerModal} title="Confirm delete" size="sm">
	<p>This cannot be undone.</p>
	{#snippet footer()}
		<Button variant="ghost" onclick={() => (showDangerModal = false)}>Cancel</Button>
		<Button
			variant="danger"
			onclick={() => {
				showDangerModal = false;
				toast.error('Deleted');
			}}
		>
			Delete
		</Button>
	{/snippet}
</Modal>

<Toast />

<style>
	.brand {
		font-size: 13px;
		letter-spacing: 0.02em;
	}
	.navbrand {
		font-size: 12.5px;
		font-weight: 650;
		color: var(--ink-2);
	}
	.demo {
		max-width: 980px;
		margin: 0 auto;
		display: grid;
		gap: 1.4rem;
	}
	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 1rem;
		border-bottom: 1px solid var(--border);
		padding-bottom: 1rem;
	}
	h1 {
		margin: 0;
		font-size: 1.4rem;
		font-weight: 680;
	}
	.sub {
		margin: 0.3rem 0 0;
		color: var(--muted);
		font-size: 0.85rem;
	}
	.chrome {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}
	h2 {
		font-size: 0.78rem;
		font-weight: 700;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		color: var(--muted);
		margin: 0 0 0.9rem;
	}
	.row {
		display: flex;
		flex-wrap: wrap;
		gap: 0.6rem;
	}
	.gallery {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(56px, 1fr));
		gap: 6px;
	}
	.icon-tile {
		display: grid;
		place-items: center;
		aspect-ratio: 1;
		border: 1px solid var(--border);
		border-radius: 8px;
		background: var(--surface-2);
		color: var(--ink-2);
	}
	.icon-tile:hover {
		background: var(--hover);
		color: var(--ink);
	}
	.field {
		display: grid;
		gap: 4px;
		font-size: 12px;
		color: var(--muted);
	}
</style>
