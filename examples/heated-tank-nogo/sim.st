PROGRAM Sim
(* The plant itself, simulated in IEC — so this example runs, moves, and
   alarms with zero Go. The pump fills the tank against a constant demand;
   the heater warms it against ambient loss. A real deployment deletes
   this task and points the driver at the field (or keeps physics in Go —
   that tier still exists; see examples/heated-tank-fbd). *)
VAR_EXTERNAL
    LevelPct : REAL;
    TempC    : REAL;
    SimDtS   : REAL;
    PumpRun  : BOOL;
    Heater   : REAL;
END_VAR

(* tank: pump inflow vs demand, in %/s *)
IF PumpRun THEN
    LevelPct := LIMIT(0.0, LevelPct + 1.7 * SimDtS, 100.0);
ELSE
    LevelPct := LIMIT(0.0, LevelPct - 1.5 * SimDtS, 100.0);
END_IF;

(* thermal: heater warms, ambient pulls toward 20 °C (equilibrium ~80 °C
   at full heat, so the PI has headroom around the 65 °C setpoint) *)
TempC := LIMIT(0.0, TempC + (Heater * 0.012 - (TempC - 20.0) * 0.02) * SimDtS, 110.0);
END_PROGRAM
