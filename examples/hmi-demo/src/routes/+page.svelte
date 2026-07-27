<script lang="ts">
	// A nautilus operator screen built entirely from the HMI kit. It reads one
	// SSE stream (`/api/stream`, proxied to the controller) and renders the
	// tank faceplate, alarm pills, trends, setpoint controls, the field-driver
	// status cards, and the scan diagnostics — the whole kit against a live
	// controller. Point it at any nautilus runtime via CONTROLLER_URL.
	import { onMount } from 'svelte';
	import {
		Trend,
		StatusPill,
		NumberField,
		ScanDiagnostics,
		DriverStatusPanel,
		ThemeSwitch,
		Mimic,
		Faceplate,
		Tabs,
		Gauge,
		createRealtimeClient,
		TrendBuffer,
		type NautilusFrame,
		type MimicDoc,
		type MimicEquipment
	} from '@joyautomation/nautilus-hmi';
	import mimicDoc from './heated-tank.mimic.json';
	// A custom, app-supplied component (not part of the kit) — passed to
	// <Mimic> via `registry` and placed in the doc as component "HeatExchanger".
	import HeatExchanger from '$lib/HeatExchanger.svelte';

	// Two rolling windows for the trend (3 minutes).
	const tempBuf = new TrendBuffer(180);
	const levelBuf = new TrendBuffer(180);

	const rt = createRealtimeClient<NautilusFrame>({
		url: '/api/stream',
		onFrame: (f) => {
			const t = f.tags ?? {};
			if (typeof t.TempC === 'number') tempBuf.push(f.ts, t.TempC);
			if (typeof t.LevelPct === 'number') levelBuf.push(f.ts, t.LevelPct);
		}
	});

	onMount(() => {
		rt.start();
		return () => rt.stop();
	});

	// Tag accessors over the latest frame, with sane fallbacks.
	const tags = $derived((rt.frame?.tags ?? {}) as Record<string, unknown>);
	const num = (k: string, d = 0) => (typeof tags[k] === 'number' ? (tags[k] as number) : d);
	const bool = (k: string) => tags[k] === true;
	const str = (k: string) => (typeof tags[k] === 'string' ? (tags[k] as string) : '');

	// Write a tag back to the controller (same-origin via the dev proxy, so
	// the write is allowed without a token).
	async function writeTag(name: string, value: unknown) {
		try {
			await fetch('/api/tags', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ name, value })
			});
		} catch {
			/* offline — the next frame will show the unchanged value */
		}
	}

	// Alarm pill state.
	const hiAlm = $derived(bool('HiTempAlm'));
	const lowAlm = $derived(bool('TempLowAlm'));
	const horn = $derived(bool('Horn'));

	// This screen is drawn for the heated-tank program. When the connected
	// controller doesn't publish those tags, say so and dim the faceplates —
	// an absent tag must never masquerade as a confident 0.0.
	const hasTankTags = $derived('TempC' in tags && 'LevelPct' in tags);

	// Faceplates: clicking equipment on the mimic opens its device popup —
	// the PLC-HMI pattern (monitoring | control, tabbed subviews).
	let faceplate = $state<string | null>(null);
	let faceplateTab = $state(0);
	function openFaceplate(eq: MimicEquipment) {
		faceplate = eq.id === 'TIC101' ? 'T101' : eq.id; // the gauge belongs to the tank
		faceplateTab = 0;
	}
</script>

<div class="app">
	<header>
		<div class="brand">
			<h1>Heated Tank <span class="pou">T-101</span></h1>
			<p class="sub">nautilus HMI kit · live from the controller's tag stream</p>
		</div>
		<div class="chrome">
			<StatusPill kind={rt.connected ? 'good' : 'critical'} label={rt.connected ? 'live' : 'stale'} />
			<a class="navlink" href="/trends">Trends</a>
			<a class="navlink" href="/primitives">Primitives</a>
			<ThemeSwitch />
		</div>
	</header>

	{#if rt.frame}
		<!-- Field connectivity first: is the plant even reachable? -->
		<section>
			<h2>Field drivers</h2>
			<DriverStatusPanel drivers={rt.frame.drivers ?? []} />
		</section>

		{#if !hasTankTags}
			<div class="notank" role="status">
				This controller doesn't publish the heated-tank tags (<code>TempC</code>,
				<code>LevelPct</code>, …) — the faceplates below are idle, not live zeros.
				Run <code>examples/heated-tank-nogo</code> and point <code>CONTROLLER_URL</code>
				at it to see them move.
			</div>
		{/if}

		<section class:idle={!hasTankTags}>
			<h2>Process</h2>
			<!-- The P&ID mimic: pure data (heated-tank.mimic.json) rendered
			     live by <Mimic> — equipment placement, pipe routing, and tag
			     bindings all live in the document, not this page. -->
			<Mimic
				doc={mimicDoc as unknown as MimicDoc}
				{tags}
				registry={{ HeatExchanger }}
				onequipmentclick={openFaceplate}
			/>
			<div class="plantfoot">
				<div class="alarms">
					<StatusPill kind={horn ? 'critical' : 'off'} label={horn ? 'HORN' : 'horn clear'} />
					<StatusPill kind={hiAlm ? 'critical' : 'off'} label="hi temp" />
					<StatusPill kind={lowAlm ? 'warning' : 'off'} label="lo temp" />
					<StatusPill kind={bool('PumpRun') ? 'good' : 'off'} label={bool('PumpRun') ? 'pump run' : 'pump off'} />
				</div>
				{#if str('Status')}<p class="statusline">{str('Status')}</p>{/if}
			</div>
		</section>

		<section class:idle={!hasTankTags}>
			<h2>Trends</h2>
			<Trend
				series={[
					{ name: 'Temp', color: 'var(--s1)', points: tempBuf.points },
					{ name: 'Level', color: 'var(--s2)', points: levelBuf.points }
				]}
				unit=""
			/>
		</section>

		<section class:idle={!hasTankTags}>
			<h2>Setpoints</h2>
			<div class="controls">
				<NumberField
					label="Temp setpoint"
					unit="°C"
					value={num('TempSP', 65)}
					min={0}
					max={100}
					step={0.5}
					onsubmit={(v) => writeTag('TempSP', v)}
				/>
				<NumberField
					label="Pump start level"
					unit="%"
					value={num('PumpStartLevel', 40)}
					min={0}
					max={100}
					onsubmit={(v) => writeTag('PumpStartLevel', v)}
				/>
				<button class="btn" class:on={bool('EcoMode')} onclick={() => writeTag('EcoMode', !bool('EcoMode'))}>
					Eco mode {bool('EcoMode') ? 'on' : 'off'}
				</button>
				<button class="btn ack" disabled={!horn} onclick={() => writeTag('HornAck', true)}>
					Acknowledge horn
				</button>
			</div>
		</section>

		<section>
			<h2>Scan diagnostics</h2>
			<ScanDiagnostics scan={rt.frame.scan} />
		</section>
	{:else}
		<div class="waiting">
			<p>Connecting to the controller…</p>
			<p class="hint">
				Run one with <code>nautilus run</code> (e.g. in <code>examples/heated-tank-nogo</code>),
				then this page follows its <code>/api/stream</code>.
			</p>
		</div>
	{/if}
</div>

{#if faceplate === 'T101'}
	<Faceplate title="T-101 · Heated surge tank" subtitle="Main/T101" onclose={() => (faceplate = null)}>
		<div class="fp-cols">
			<div class="fp-panel">
				<span class="fp-head">Monitoring</span>
				<div class="fp-rows">
					<div class="fp-row"><span>Level</span><b>{num('LevelPct').toFixed(1)} %</b></div>
					<div class="fp-row"><span>Temperature</span><b>{num('TempC').toFixed(1)} °C</b></div>
					<div class="fp-row"><span>Heater output</span><b>{num('Heater').toFixed(0)} %</b></div>
					<div class="fp-row"><span>Rate of change</span><b>{num('TempRate').toFixed(2)} °C/min</b></div>
				</div>
				<div class="fp-pills">
					<StatusPill kind={hiAlm ? 'critical' : 'off'} label="hi temp" />
					<StatusPill kind={lowAlm ? 'warning' : 'off'} label="lo temp" />
					<StatusPill kind={horn ? 'critical' : 'off'} label="horn" />
				</div>
			</div>
			<div class="fp-panel">
				<span class="fp-head">Control</span>
				<NumberField
					label="Temp setpoint"
					unit="°C"
					value={num('TempSP', 65)}
					min={0}
					max={100}
					step={0.5}
					onsubmit={(v) => writeTag('TempSP', v)}
				/>
				<button class="btn" class:on={bool('EcoMode')} onclick={() => writeTag('EcoMode', !bool('EcoMode'))}>
					Eco mode {bool('EcoMode') ? 'on' : 'off'}
				</button>
				<button class="btn ack" disabled={!horn} onclick={() => writeTag('HornAck', true)}>
					Acknowledge horn
				</button>
			</div>
		</div>
		<Tabs tabs={['Trend', 'Tuning']} bind:active={faceplateTab} />
		{#if faceplateTab === 0}
			<Trend
				series={[{ name: 'Temp', color: 'var(--s1)', points: tempBuf.points }]}
				unit="°C"
				height={150}
			/>
		{:else}
			<div class="fp-tuning">
				<NumberField label="Kp" value={num('Kp', 12)} min={0} max={100} step={0.5} onsubmit={(v) => writeTag('Kp', v)} />
				<NumberField label="Ki" unit="1/s" value={num('Ki', 0.15)} min={0} max={10} step={0.01} onsubmit={(v) => writeTag('Ki', v)} />
				<NumberField label="Eco setpoint" unit="°C" value={num('TempSPEco', 55)} min={0} max={100} step={0.5} onsubmit={(v) => writeTag('TempSPEco', v)} />
			</div>
		{/if}
	</Faceplate>
{:else if faceplate === 'P101'}
	<Faceplate title="P-101 · Feed pump" subtitle="Main/P101" onclose={() => (faceplate = null)}>
		<div class="fp-cols">
			<div class="fp-panel">
				<span class="fp-head">Monitoring</span>
				<div class="fp-center">
					<Gauge value={num('LevelPct')} min={0} max={100} unit="%" label="Tank level" width={150} />
				</div>
				<div class="fp-pills">
					<StatusPill kind={bool('PumpRun') ? 'good' : 'off'} label={bool('PumpRun') ? 'running' : 'stopped'} />
				</div>
			</div>
			<div class="fp-panel">
				<span class="fp-head">Control</span>
				<p class="fp-note">Seal-in latch: starts at the low mark, drops out at the high mark.</p>
				<NumberField
					label="Start level"
					unit="%"
					value={num('PumpStartLevel', 40)}
					min={0}
					max={100}
					onsubmit={(v) => writeTag('PumpStartLevel', v)}
				/>
				<NumberField
					label="Stop level"
					unit="%"
					value={num('PumpStopLevel', 75)}
					min={0}
					max={100}
					onsubmit={(v) => writeTag('PumpStopLevel', v)}
				/>
			</div>
		</div>
	</Faceplate>
{/if}

<style>
	.app {
		max-width: 1080px;
		margin: 0 auto;
		padding: 1.5rem 1.25rem 3rem;
		display: grid;
		gap: 1.6rem;
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
		font-size: 1.5rem;
		font-weight: 680;
	}
	h1 .pou {
		font-family: var(--mono);
		font-weight: 600;
		font-size: 1rem;
		color: var(--muted);
		margin-left: 0.4rem;
	}
	.sub {
		margin: 0.3rem 0 0;
		color: var(--muted);
		font-size: 0.85rem;
	}
	.chrome {
		display: flex;
		align-items: center;
		gap: 0.6rem;
	}
	.navlink {
		border: 1px solid var(--border);
		background: var(--surface);
		color: var(--ink-2);
		border-radius: var(--radius, 8px);
		padding: 0.45rem 0.8rem;
		font-size: 0.82rem;
		font-weight: 550;
	}
	.navlink:hover {
		border-color: var(--accent);
		color: var(--ink);
	}
	h2 {
		font-size: 0.78rem;
		font-weight: 700;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		color: var(--muted);
		margin: 0 0 0.7rem;
	}
	.plantfoot {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 1rem;
		margin-top: 0.6rem;
	}
	.alarms {
		display: flex;
		flex-wrap: wrap;
		gap: 0.4rem;
	}
	.statusline {
		margin: 0;
		font-family: var(--mono);
		font-size: 0.8rem;
		color: var(--ink-2);
	}
	.controls {
		display: flex;
		flex-wrap: wrap;
		gap: 0.8rem;
		align-items: flex-end;
	}
	.btn {
		border: 1px solid var(--border);
		background: var(--surface);
		color: var(--ink);
		border-radius: var(--radius, 8px);
		padding: 0.55rem 0.9rem;
		font-size: 0.88rem;
		font-weight: 550;
		cursor: pointer;
	}
	.btn:hover {
		border-color: var(--accent);
	}
	.btn.on {
		border-color: var(--good);
		color: var(--good);
	}
	.btn.ack:disabled {
		opacity: 0.45;
		cursor: default;
	}
	/* faceplate content layout */
	.fp-cols {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 12px;
	}
	.fp-panel {
		border: 1px solid var(--border);
		border-radius: 8px;
		padding: 10px 12px;
		display: grid;
		gap: 10px;
		align-content: start;
		min-width: 0;
	}
	.fp-head {
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--muted);
	}
	.fp-rows {
		display: grid;
		gap: 6px;
	}
	.fp-row {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		font-size: 12.5px;
		color: var(--ink-2);
	}
	.fp-row b {
		font-family: var(--mono);
		font-weight: 650;
		font-variant-numeric: tabular-nums;
		color: var(--ink);
	}
	.fp-pills {
		display: flex;
		flex-wrap: wrap;
		gap: 4px;
	}
	.fp-center {
		display: grid;
		justify-items: center;
	}
	.fp-note {
		margin: 0;
		font-size: 11.5px;
		color: var(--muted);
	}
	.fp-tuning {
		display: flex;
		flex-wrap: wrap;
		gap: 10px;
	}
	@media (max-width: 480px) {
		.fp-cols {
			grid-template-columns: 1fr;
		}
	}
	.notank {
		background: var(--surface);
		border: 1px solid color-mix(in srgb, var(--warn) 45%, var(--border));
		border-radius: var(--radius, 8px);
		padding: 0.7rem 0.9rem;
		font-size: 0.85rem;
		color: var(--ink-2);
	}
	.idle {
		opacity: 0.45;
	}
	.waiting {
		padding: 3rem 1rem;
		text-align: center;
		color: var(--muted);
	}
	.waiting .hint {
		font-size: 0.85rem;
	}
	code {
		font-family: var(--mono);
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: 4px;
		padding: 0.05rem 0.3rem;
		color: var(--accent);
	}
	@media (max-width: 720px) {
		}
</style>
