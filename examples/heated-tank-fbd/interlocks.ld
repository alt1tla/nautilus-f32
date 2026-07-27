PROGRAM Interlocks
(* The annunciator logic as LADDER — a third language on the same tag
   store, its own 200 ms task (see Tasks in main.go). Open the ladder:
   VS Code -> "nautilus: Open Ladder Diagram Preview". With live values
   on, power flow paints the rungs green as the plant runs.

   Rung grammar: series = AND, [ a | b ] = OR branch, /x = NC contact,
   ( x ) coil, ( S x )/( R x ) latch/unlatch, FN(args) as a contact,
   inst:TYPE(args) puts a function block in the rung. *)
VAR_EXTERNAL
    TempLowAlm : BOOL;
    TempC      : REAL;
    HornAck    : BOOL;
    Horn       : BOOL;
    HiTempAlm  : BOOL;
END_VAR
LD
  RUNG hitemp (* comparison as a contact, with a 5 s on-delay *)
    GT(TempC, 90.0) t2:TON(PT := T#5S) ( HiTempAlm )

  // Comment runs like this one render as notes in the ladder diagram —
  // dblclick to edit them there, empty text deletes.
  RUNG horn (* either alarm sounds the horn until the operator acks *)
    [ TempLowAlm | HiTempAlm ] /HornAck ( Horn )

  RUNG ackclear (* the ack releases itself once both alarms clear *)
    /TempLowAlm /HiTempAlm ( R HornAck )
END_LD
END_PROGRAM
