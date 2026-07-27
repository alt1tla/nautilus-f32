<script lang="ts">
	// Showcase for TrendChart: a live multi-pen process trend (simulated —
	// no controller needed), a shared-units variant of the same signals, and
	// a static/historical batch-cycle trend with a sensor-dropout gap.
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import {
		AppShell,
		MenuBar,
		Nav,
		Icon,
		theme,
		TrendChart,
		TrendBuffer,
		type MenuBarMenu,
		type NavSection,
		type TrendPen,
		type TrendThreshold,
		type TrendPoint
	} from '@joyautomation/nautilus-hmi';

	const menus: MenuBarMenu[] = [
		{
			label: 'View',
			items: [
				{ label: 'Heated tank', icon: 'fire', href: '/' },
				{ label: 'Primitives', icon: 'cog-6-tooth', href: '/primitives' }
			]
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
		}
	];

	// ── Simulated process: a heater loop chasing a setpoint, with a valve
	// crudely PI-controlling it — realistic-enough lag + noise to make a
	// trend worth watching, no controller required. ──────────────────────────
	const WINDOW_MS = 4 * 60_000; // 4-minute rolling window
	const tempBuf = new TrendBuffer(600); // keep 10 min of history around
	const spBuf = new TrendBuffer(600);
	const valveBuf = new TrendBuffer(600);

	let tempC = $state(58);
	let setpoint = $state(65);
	let valvePct = $state(35);
	let paused = $state(false);

	let integral = 0;
	function simStep(t: number) {
		const error = setpoint - tempC;
		integral = Math.max(-15, Math.min(15, integral + error * 0.02));
		const control = error * 1.1 + integral * 0.06;
		valvePct = Math.max(0, Math.min(100, 38 + control * 3.2 + (Math.random() - 0.5) * 3));
		const target = 18 + valvePct * 0.72;
		tempC += (target - tempC) * 0.035 + (Math.random() - 0.5) * 0.35;

		tempBuf.push(t, tempC);
		spBuf.push(t, setpoint);
		valveBuf.push(t, valvePct);
	}

	// Occasional operator setpoint changes, so the loop has something to chase.
	function bumpSetpoint() {
		setpoint = Math.round((55 + Math.random() * 25) * 2) / 2;
	}

	let simTimer: ReturnType<typeof setInterval>;
	let spTimer: ReturnType<typeof setInterval>;
	onMount(() => {
		// Seed a full window of backdated history so the chart isn't empty (or
		// half-empty) on first paint — each seed step gets its own past
		// timestamp rather than clustering at "now".
		const seedSteps = Math.ceil(WINDOW_MS / 500);
		const seedStart = Date.now() - seedSteps * 500;
		for (let i = 0; i < seedSteps; i++) simStep(seedStart + i * 500);
		simTimer = setInterval(() => simStep(Date.now()), 500);
		spTimer = setInterval(bumpSetpoint, 18_000);
		return () => {
			clearInterval(simTimer);
			clearInterval(spTimer);
		};
	});

	// Pens for the percent-normalized view: °C and % share nothing, so each
	// pen is normalized to its own range instead of forcing a dual axis.
	const percentPens = $derived<TrendPen[]>([
		{ id: 'temp', label: 'Temperature', units: '°C', min: 20, max: 95, data: tempBuf.points },
		{ id: 'sp', label: 'Setpoint', units: '°C', dashed: true, min: 20, max: 95, data: spBuf.points },
		{ id: 'valve', label: 'Valve position', units: '%', min: 0, max: 100, data: valveBuf.points }
	]);

	const thresholds: TrendThreshold[] = [
		{ penId: 'temp', kind: 'warn', lo: 80, hi: 90, label: 'high temp warn' },
		{ penId: 'temp', kind: 'crit', lo: 90, hi: 95, label: 'high temp crit' }
	];

	// Same two temperature pens, sharing one engineering-units axis — the
	// simpler mode, appropriate because both pens are °C.
	const sharedPens = $derived<TrendPen[]>([
		{ id: 'temp2', label: 'Temperature', units: '°C', data: tempBuf.points },
		{ id: 'sp2', label: 'Setpoint', units: '°C', dashed: true, data: spBuf.points }
	]);
	const sharedThresholds: TrendThreshold[] = [
		{ penId: 'temp2', kind: 'warn', value: 80, label: 'warn' },
		{ penId: 'temp2', kind: 'crit', value: 90, label: 'crit' }
	];

	// ── Static/historical example: a finished batch cycle, generated once,
	// with a deliberate ~7-minute sensor-dropout gap to show gap-breaking. ──
	function genBatch(): TrendPoint[] {
		const now = Date.now();
		const pts: TrendPoint[] = [];
		const totalMin = 90;
		for (let m = 0; m <= totalMin; m += 0.5) {
			if (m > 38 && m < 45) continue; // dropout
			const phase = m < 20 ? m / 20 : m < 60 ? 1 : Math.max(0, 1 - (m - 60) / 25);
			const v = 22 + phase * 58 + (Math.random() - 0.5) * 1.1;
			pts.push({ t: now - (totalMin - m) * 60_000, v });
		}
		return pts;
	}
	const batchPens: TrendPen[] = [{ id: 'batch', label: 'Batch temp', units: '°C', color: 'var(--s1)', data: genBatch() }];
	const batchThresholds: TrendThreshold[] = [{ penId: 'batch', kind: 'warn', lo: 75, hi: 82, label: 'target band' }];
</script>

<svelte:head>
	<title>nautilus HMI · trends</title>
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
		<Nav sections={navSections} current={page.url.pathname}>
			{#snippet brand()}<span class="navbrand">HMI primitives</span>{/snippet}
		</Nav>
	{/snippet}

	<div class="demo">
		<header>
			<div>
				<h1>Trends</h1>
				<p class="sub">TrendChart — multi-pen, live + historical, alarm bands, pause/resume, legend hide/show</p>
			</div>
			<button class="btn" onclick={() => theme.set(theme.effective === 'dark' ? 'light' : 'dark')}>
				<Icon name={theme.effective === 'dark' ? 'sun' : 'moon'} size={16} />
			</button>
		</header>

		<section class="card">
			<h2>Live process · normalized (% of range)</h2>
			<p class="note">
				Temperature, its setpoint, and valve position — °C and % share nothing, so each pen is
				normalized to its own range rather than forcing a second y-axis. Setpoint bumps every ~18s
				so the loop has something to chase. Click a pen below to hide/show it; the pause button
				freezes the view while data keeps accumulating underneath.
			</p>
			<TrendChart pens={percentPens} thresholds={thresholds} axisMode="percent" windowMs={WINDOW_MS} bind:paused />
			<p class="status">Chart reports: <b>{paused ? 'paused' : 'live'}</b></p>
		</section>

		<section class="card">
			<h2>Live process · shared axis (same units)</h2>
			<p class="note">The same temperature + setpoint pair, sharing one auto-scaled °C axis — the simpler mode, used here because both pens are already the same unit.</p>
			<TrendChart pens={sharedPens} thresholds={sharedThresholds} axisMode="shared" windowMs={WINDOW_MS} height={200} />
		</section>

		<section class="card">
			<h2>Historical · static batch record</h2>
			<p class="note">A finished batch, loaded as a plain array (no live window) — the chart auto-fits the full span. The ~7-minute gap in the middle is a simulated sensor dropout, rendered as a break rather than a straight line across it (<code>gapMs</code>).</p>
			<TrendChart pens={batchPens} thresholds={batchThresholds} height={200} gapMs={4 * 60_000} />
		</section>
	</div>
</AppShell>

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
		max-width: 900px;
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
	h2 {
		font-size: 0.78rem;
		font-weight: 700;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		color: var(--muted);
		margin: 0 0 0.5rem;
	}
	.note {
		margin: 0 0 0.9rem;
		font-size: 0.82rem;
		color: var(--ink-2);
		max-width: 68ch;
	}
	.status {
		margin: 0.6rem 0 0;
		font-size: 0.78rem;
		color: var(--muted);
	}
	.status b {
		color: var(--ink);
	}
	code {
		font-family: var(--mono);
		background: var(--surface-2);
		border: 1px solid var(--border);
		border-radius: 4px;
		padding: 0.05rem 0.3rem;
		color: var(--accent);
	}
	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
	}
</style>
