PROGRAM AllSupportedFeatures

VAR
    x : REAL;
    y : REAL;
    z : REAL;
    selector : REAL;
    bitIndex : REAL;
    boolA : BOOL;
    boolB : BOOL;
    boolResult : BOOL;
    intResult : INT;
    timeResult : TIME;
    text : STRING;
    textResult : STRING;

    ton1 : TON;
    tof1 : TOF;
    tp1 : TP;
    rising : R_TRIG;
    falling : F_TRIG;
    counterUp : CTU;
    counterDown : CTD;
    counterBoth : CTUD;
    setLatch : SR;
    resetLatch : RS;
END_VAR

(* Target value. The embedding compiler replaces VAL with a controller read. *)
x := VAL('Controller1.Temperature');
y := 5;
selector := 1;
bitIndex := 31;
boolA := x > 0;
boolB := y <= 10;

(* Arithmetic and comparison operators. *)
z := x + y;
z := z - 1;
z := z * 2;
z := z / 3;
z := z MOD 5;
boolResult := x = y;
boolResult := x <> y;
boolResult := x > y;
boolResult := x >= y;
boolResult := x < y;
boolResult := x <= y;
boolResult := boolA AND boolB;
boolResult := boolA OR boolB;
boolResult := boolA XOR boolB;
boolResult := NOT boolA;

(* The logical and remainder operators are also accepted as functions. *)
boolResult := AND(boolA, boolB, TRUE);
boolResult := OR(boolA, boolB, FALSE);
boolResult := XOR(boolA, boolB, TRUE);
z := MOD(x, y);

(* Numeric functions. *)
z := ABS(-x);
z := MIN(x, y, 100);
z := MAX(x, y, -100);
z := LIMIT(-50, x, 50);
z := SQRT(16);
z := LN(2.7182818);
z := LOG(1000);
z := EXP(1);
z := EXPT(2, 8);
z := SIN(0.5);
z := COS(0.5);
z := TAN(0.5);
z := ASIN(0.5);
z := ACOS(0.5);
z := ATAN(1);
z := ATAN2(1, 1);
intResult := TRUNC(-3.75);

(* Selection functions. *)
z := SEL(boolA, x, y);
z := MUX(selector, 10, 20, 30);

(* Integer bit operations and the float32 bit bridge. *)
intResult := SHL(3, 2);
intResult := SHR(12, 2);
intResult := ROL(1, 8);
intResult := ROR(256, 8);
boolResult := FTB(x, bitIndex);
z := BTF(
    FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,
    FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,
    FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE,
    FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE, FALSE
);

(* Scalar and string conversions. Numeric values are float32-oriented in
   the default analysis mode, but conversion names remain available. *)
z := INT_TO_REAL(intResult);
intResult := REAL_TO_INT(z);
intResult := BOOL_TO_INT(boolA);
boolResult := INT_TO_BOOL(intResult);
timeResult := INT_TO_TIME(intResult);
intResult := TIME_TO_INT(timeResult);
timeResult := REAL_TO_TIME(z);
z := TIME_TO_REAL(timeResult);
z := BOOL_TO_REAL(boolA);
boolResult := REAL_TO_BOOL(z);

text := 'Structured Text';
intResult := LEN(text);
textResult := LEFT(text, 10);
textResult := RIGHT(text, 4);
textResult := MID(text, 6, 2);
textResult := CONCAT('PLC', '-', 'TEST');
textResult := INSERT('ABCD', 'XY', 2);
textResult := DELETE('ABXYCD', 2, 3);
textResult := REPLACE('ABCDE', 'XY', 2, 2);
intResult := FIND(text, 'Text');
textResult := INT_TO_STRING(intResult);
textResult := REAL_TO_STRING(z);
textResult := BOOL_TO_STRING(boolA);
textResult := TIME_TO_STRING(timeResult);
intResult := STRING_TO_INT('42');
z := STRING_TO_REAL('12.5');
boolResult := STRING_TO_BOOL('TRUE');

(* Every built-in function block is declared once and called once. Numeric
   timer values are milliseconds: PT := 2000 means two seconds. *)
ton1(IN := boolA, PT := 2000);
tof1(IN := boolA, PT := 2000);
tp1(IN := boolA, PT := 500);
rising(CLK := boolA);
falling(CLK := boolA);
counterUp(CU := rising.Q, R := FALSE, PV := 10);
counterDown(CD := falling.Q, LD := FALSE, PV := 10);
counterBoth(CU := rising.Q, CD := falling.Q, R := FALSE, LD := FALSE, PV := 10);
setLatch(S1 := rising.Q, R := falling.Q);
resetLatch(S := rising.Q, R1 := falling.Q);

(* Read every FB output so member access is checked as well. *)
boolResult := ton1.Q OR tof1.Q OR tp1.Q;
timeResult := ton1.ET + tof1.ET + tp1.ET;
boolResult := rising.Q OR falling.Q;
boolResult := counterUp.Q OR counterDown.Q;
intResult := counterUp.CV + counterDown.CV + counterBoth.CV;
boolResult := counterBoth.QU OR counterBoth.QD;
boolResult := setLatch.Q1 OR resetLatch.Q1;

(* Control-flow structures understood by the ST parser/lowerer. *)
IF boolA THEN
    z := z + 1;
ELSIF boolB THEN
    z := z + 2;
ELSE
    z := 0;
END_IF;

CASE selector OF
    0: z := 10;
    1, 2: z := 20;
ELSE
    z := 30;
END_CASE;

WHILE z < 25 DO
    z := z + 1;
    IF z > 22 THEN
        EXIT;
    END_IF;
END_WHILE;

REPEAT
    z := z - 1;
UNTIL z <= 20
END_REPEAT;

FOR intResult := 0 TO 3 BY 1 DO
    z := z + intResult;
END_FOR;

END_PROGRAM
