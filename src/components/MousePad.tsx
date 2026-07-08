import React, { useEffect, useRef, useState } from "react";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Grid from "@mui/material/Grid";
import Slider from "@mui/material/Slider";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import ToggleButton from "@mui/material/ToggleButton";
import ToggleButtonGroup from "@mui/material/ToggleButtonGroup";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import PanToolIcon from "@mui/icons-material/PanTool";
import LockIcon from "@mui/icons-material/Lock";

import { MouseButtons } from "../lib/models";

type MouseMode = "relative" | "absolute";

// movement (in CSS pixels) below which a press/release is treated as a tap
const tapThreshold = 10;
// maximum duration (ms) of a tap
const tapDuration = 300;
// max time (ms) between two taps to register as a double-tap
const doubleTapWindow = 250;
// max distance (px) between two taps to register as a double-tap
const doubleTapDist = 24;
// how long to hold still before a press becomes a click-and-hold drag
const holdToDragDelay = 400;

// map a DOM mouse button number to its name
function buttonName(button: number): string | null {
  if (button === 0) return "left";
  if (button === 1) return "middle";
  if (button === 2) return "right";
  return null;
}

interface MousePadProps {
  open: boolean;
  onClose: () => void;
  sendMessage: (message: string) => void;
}

export default function MousePad(props: MousePadProps) {
  const { open, onClose, sendMessage } = props;

  const [mode, setMode] = useState<MouseMode>("relative");
  const [sensitivity, setSensitivity] = useState<number>(1.5);
  // dragLock: left button held down via the Hold button, until toggled off
  const [dragLock, setDragLock] = useState<boolean>(false);
  // held: left button currently down by any means (lock or long-press gesture)
  const [held, setHeld] = useState<boolean>(false);
  // captured: pointer lock is active (desktop mouse capture)
  const [captured, setCaptured] = useState<boolean>(false);
  // whether the browser supports pointer lock at all
  const [canCapture] = useState<boolean>(
    () =>
      typeof document !== "undefined" &&
      "pointerLockElement" in document &&
      typeof HTMLElement !== "undefined" &&
      "requestPointerLock" in HTMLElement.prototype
  );

  const padRef = useRef<HTMLDivElement | null>(null);
  const lastPointRef = useRef<{ x: number; y: number } | null>(null);
  const movedRef = useRef<number>(0);
  const downTimeRef = useRef<number>(0);
  const pendingRef = useRef<{ dx: number; dy: number }>({ dx: 0, dy: 0 });
  const rafRef = useRef<number | null>(null);
  const lastTapTimeRef = useRef<number | null>(null);
  const lastTapPosRef = useRef<{ x: number; y: number } | null>(null);
  const clickTimerRef = useRef<number | null>(null);
  const holdTimerRef = useRef<number | null>(null);
  // dragLockRef / gestureHeldRef mirror the two ways the button can be held,
  // readable synchronously from pointer handlers
  const dragLockRef = useRef<boolean>(false);
  const gestureHeldRef = useRef<boolean>(false);
  // pressActiveRef is true between pointer down and up (finger/button down);
  // when false, a pointer move is a hover (mouse over the pad, no button)
  const pressActiveRef = useRef<boolean>(false);
  // pointer-lock state readable from handlers, and the physical buttons held
  // down while captured (so we can release them if the lock is lost)
  const capturedRef = useRef<boolean>(false);
  const heldButtonsRef = useRef<Set<number>>(new Set());

  // keep the latest mode/sensitivity readable from pointer handlers
  const modeRef = useRef<MouseMode>(mode);
  const sensitivityRef = useRef<number>(sensitivity);
  useEffect(() => {
    modeRef.current = mode;
  }, [mode]);
  useEffect(() => {
    sensitivityRef.current = sensitivity;
  }, [sensitivity]);

  // always-current sendMessage for use in cleanup closures
  const sendRef = useRef(sendMessage);
  useEffect(() => {
    sendRef.current = sendMessage;
  }, [sendMessage]);

  const syncHeld = () => {
    setHeld(dragLockRef.current || gestureHeldRef.current);
  };

  const clearHoldTimer = () => {
    if (holdTimerRef.current !== null) {
      window.clearTimeout(holdTimerRef.current);
      holdTimerRef.current = null;
    }
  };

  // release any physical buttons passed through while captured
  const releasePassthrough = () => {
    heldButtonsRef.current.forEach((b) => {
      const name = buttonName(b);
      if (name) {
        sendRef.current(`mouseBtn:${name}_up`);
      }
    });
    heldButtonsRef.current.clear();
  };

  // release any held button and reset transient state
  const releaseAll = () => {
    clearHoldTimer();
    if (rafRef.current !== null) {
      window.cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
      pendingRef.current = { dx: 0, dy: 0 };
    }
    if (clickTimerRef.current !== null) {
      window.clearTimeout(clickTimerRef.current);
      clickTimerRef.current = null;
    }
    if (dragLockRef.current || gestureHeldRef.current) {
      sendRef.current("mouseBtn:left_up");
    }
    releasePassthrough();
    if (typeof document !== "undefined" && document.pointerLockElement) {
      document.exitPointerLock();
    }
    dragLockRef.current = false;
    gestureHeldRef.current = false;
    pressActiveRef.current = false;
    capturedRef.current = false;
    lastTapTimeRef.current = null;
    lastTapPosRef.current = null;
    lastPointRef.current = null;
  };

  // release everything when the dialog closes
  useEffect(() => {
    if (!open) {
      releaseAll();
      setDragLock(false);
      setHeld(false);
      setCaptured(false);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  // release everything if the component unmounts while a button is held
  useEffect(() => {
    return () => {
      releaseAll();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // track pointer-lock engage/release (Esc releases it via the browser)
  useEffect(() => {
    const onChange = () => {
      const locked =
        typeof document !== "undefined" &&
        document.pointerLockElement === padRef.current;
      capturedRef.current = locked;
      setCaptured(locked);
      if (!locked) {
        releasePassthrough();
      }
    };
    document.addEventListener("pointerlockchange", onChange);
    return () => document.removeEventListener("pointerlockchange", onChange);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // while captured, pass the physical mouse straight through: relative movement
  // and raw button down/up, so clicks, double-clicks and drags all behave like
  // a real mouse
  useEffect(() => {
    if (!captured) {
      return;
    }

    const onMove = (e: MouseEvent) => {
      pendingRef.current.dx += e.movementX * sensitivityRef.current;
      pendingRef.current.dy += e.movementY * sensitivityRef.current;
      scheduleFlush();
    };
    const onDown = (e: MouseEvent) => {
      const name = buttonName(e.button);
      if (!name) return;
      e.preventDefault();
      heldButtonsRef.current.add(e.button);
      sendRef.current(`mouseBtn:${name}_down`);
    };
    const onUp = (e: MouseEvent) => {
      const name = buttonName(e.button);
      if (!name) return;
      e.preventDefault();
      heldButtonsRef.current.delete(e.button);
      sendRef.current(`mouseBtn:${name}_up`);
    };
    const onContext = (e: Event) => e.preventDefault();

    document.addEventListener("mousemove", onMove);
    document.addEventListener("mousedown", onDown);
    document.addEventListener("mouseup", onUp);
    document.addEventListener("contextmenu", onContext);
    return () => {
      document.removeEventListener("mousemove", onMove);
      document.removeEventListener("mousedown", onDown);
      document.removeEventListener("mouseup", onUp);
      document.removeEventListener("contextmenu", onContext);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [captured]);

  const flushMove = () => {
    rafRef.current = null;
    const { dx, dy } = pendingRef.current;
    pendingRef.current = { dx: 0, dy: 0 };
    const rx = Math.round(dx);
    const ry = Math.round(dy);
    if (rx !== 0 || ry !== 0) {
      sendMessage(`mouseMove:${rx},${ry}`);
    }
  };

  const scheduleFlush = () => {
    if (rafRef.current === null) {
      rafRef.current = window.requestAnimationFrame(flushMove);
    }
  };

  const requestCapture = () => {
    const pad = padRef.current;
    if (!pad || typeof pad.requestPointerLock !== "function") {
      return;
    }
    const req = pad.requestPointerLock() as unknown as
      | Promise<void>
      | undefined;
    if (req && typeof req.catch === "function") {
      req.catch(() => {
        /* lock can fail if not from a user gesture; ignore */
      });
    }
  };

  const sendAbsolute = (clientX: number, clientY: number) => {
    const pad = padRef.current;
    if (!pad) {
      return;
    }

    const rect = pad.getBoundingClientRect();
    if (rect.width <= 0 || rect.height <= 0) {
      return;
    }

    let fx = (clientX - rect.left) / rect.width;
    let fy = (clientY - rect.top) / rect.height;
    fx = Math.min(1, Math.max(0, fx));
    fy = Math.min(1, Math.max(0, fy));

    // send position as per-mille (0-1000) of the screen; the server maps it
    // onto the pointer range, so it's independent of the video resolution
    const x = Math.round(fx * 1000);
    const y = Math.round(fy * 1000);
    sendMessage(`mousePos:${x},${y}`);
  };

  // establish a baseline when the pointer enters (e.g. a mouse hovering in) so
  // the first move doesn't jump; a real press re-baselines in handlePointerDown
  const handlePointerEnter = (e: React.PointerEvent<HTMLDivElement>) => {
    if (capturedRef.current) return;
    if (!pressActiveRef.current) {
      lastPointRef.current = { x: e.clientX, y: e.clientY };
    }
  };

  // stop hover tracking when the pointer leaves the pad (a press keeps capture)
  const handlePointerLeave = () => {
    if (capturedRef.current) return;
    if (!pressActiveRef.current) {
      lastPointRef.current = null;
    }
  };

  const handlePointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    if (capturedRef.current) return;
    e.currentTarget.setPointerCapture(e.pointerId);
    pressActiveRef.current = true;
    lastPointRef.current = { x: e.clientX, y: e.clientY };
    movedRef.current = 0;
    downTimeRef.current = e.timeStamp;

    if (modeRef.current === "absolute") {
      sendAbsolute(e.clientX, e.clientY);
    }

    // start long-press-to-drag detection (unless the drag-lock already holds
    // the button down)
    clearHoldTimer();
    if (!dragLockRef.current) {
      holdTimerRef.current = window.setTimeout(() => {
        holdTimerRef.current = null;
        gestureHeldRef.current = true;
        syncHeld();
        sendMessage("mouseBtn:left_down");
      }, holdToDragDelay);
    }
  };

  const handlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    if (capturedRef.current) return;
    const last = lastPointRef.current;
    lastPointRef.current = { x: e.clientX, y: e.clientY };
    if (!last) {
      // first sample after entering/leaving: just set the baseline
      return;
    }

    const rawDx = e.clientX - last.x;
    const rawDy = e.clientY - last.y;

    // tap / long-press bookkeeping only applies while actually pressing; a
    // hover move (mouse over the pad, no button) just moves the cursor
    if (pressActiveRef.current) {
      movedRef.current += Math.abs(rawDx) + Math.abs(rawDy);
      if (holdTimerRef.current !== null && movedRef.current >= tapThreshold) {
        clearHoldTimer();
      }
    }

    if (modeRef.current === "absolute") {
      sendAbsolute(e.clientX, e.clientY);
      return;
    }

    pendingRef.current.dx += rawDx * sensitivityRef.current;
    pendingRef.current.dy += rawDy * sensitivityRef.current;
    scheduleFlush();
  };

  const handlePointerEnd = (e: React.PointerEvent<HTMLDivElement>) => {
    if (capturedRef.current) return;
    if (!pressActiveRef.current) {
      return;
    }
    pressActiveRef.current = false;
    // keep lastPointRef so a hovering mouse keeps moving the cursor after a click

    clearHoldTimer();

    // finished a long-press drag: release the button now
    if (gestureHeldRef.current) {
      gestureHeldRef.current = false;
      syncHeld();
      sendMessage("mouseBtn:left_up");
      return;
    }

    // while the drag-lock holds the button, the pad only drags — no clicks
    if (dragLockRef.current) {
      return;
    }

    const duration = e.timeStamp - downTimeRef.current;
    const wasTap = movedRef.current < tapThreshold && duration < tapDuration;
    if (!wasTap) {
      return;
    }

    // a quick tap acts as a left click; two quick taps as a double click
    const pos = { x: e.clientX, y: e.clientY };
    const prevTime = lastTapTimeRef.current;
    const prevPos = lastTapPosRef.current;
    const isDoubleTap =
      prevTime !== null &&
      prevPos !== null &&
      e.timeStamp - prevTime < doubleTapWindow &&
      Math.abs(pos.x - prevPos.x) + Math.abs(pos.y - prevPos.y) < doubleTapDist;

    if (clickTimerRef.current !== null) {
      window.clearTimeout(clickTimerRef.current);
      clickTimerRef.current = null;
    }

    if (isDoubleTap) {
      lastTapTimeRef.current = null;
      lastTapPosRef.current = null;
      sendMessage("mouseBtn:double_click");
    } else {
      // defer the single click briefly so a following tap can upgrade it
      lastTapTimeRef.current = e.timeStamp;
      lastTapPosRef.current = pos;
      clickTimerRef.current = window.setTimeout(() => {
        clickTimerRef.current = null;
        sendMessage("mouseBtn:click");
      }, doubleTapWindow);
    }
  };

  const toggleDragLock = () => {
    if (dragLockRef.current) {
      dragLockRef.current = false;
      setDragLock(false);
      syncHeld();
      sendMessage("mouseBtn:left_up");
    } else {
      clearHoldTimer();
      dragLockRef.current = true;
      setDragLock(true);
      syncHeld();
      sendMessage("mouseBtn:left_down");
    }
  };

  const clickButton = (button: MouseButtons) => {
    sendMessage(`mouseBtn:${button}`);
  };

  const buttons: { label: string; button: MouseButtons }[] = [
    { label: "Left Click", button: "click" },
    { label: "Double Click", button: "double_click" },
    { label: "Right Click", button: "right" },
    { label: "Middle Click", button: "middle" },
  ];

  let padHint: string;
  if (held) {
    padHint = dragLock
      ? "Left button held — move to drag, then tap Release"
      : "Left button held — move to drag, lift to release";
  } else if (mode === "absolute") {
    padHint = "Move over the pad to position · tap to click";
  } else {
    padHint = "Move over the pad to move · tap = click · hold = drag";
  }

  const padActive = held || captured;

  return (
    <Dialog
      open={open}
      onClose={(_, reason) => {
        // while the mouse is captured, Esc releases the pointer lock — don't
        // let it also close the dialog
        if (reason === "escapeKeyDown" && capturedRef.current) {
          return;
        }
        onClose();
      }}
      fullWidth
    >
      <DialogContent>
        <Stack spacing={2}>
          <ToggleButtonGroup
            color="primary"
            value={mode}
            exclusive
            fullWidth
            size="small"
            disabled={captured}
            onChange={(_, value) => {
              if (value !== null) {
                setMode(value as MouseMode);
              }
            }}
          >
            <ToggleButton value="relative">Trackpad</ToggleButton>
            <ToggleButton value="absolute">Absolute</ToggleButton>
          </ToggleButtonGroup>

          <Box
            ref={padRef}
            onPointerEnter={handlePointerEnter}
            onPointerLeave={handlePointerLeave}
            onPointerDown={handlePointerDown}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerEnd}
            onPointerCancel={handlePointerEnd}
            sx={{
              height: "40vh",
              minHeight: 200,
              border: padActive ? 2 : 1,
              borderColor: padActive ? "primary.main" : "divider",
              borderRadius: 2,
              bgcolor: padActive ? "action.selected" : "action.hover",
              display: "flex",
              flexDirection: "column",
              gap: 1,
              alignItems: "center",
              justifyContent: "center",
              textAlign: "center",
              px: 2,
              touchAction: "none",
              userSelect: "none",
              cursor: captured ? "none" : mode === "absolute" ? "crosshair" : "move",
            }}
          >
            {captured ? (
              <>
                <LockIcon color="primary" />
                <Typography variant="subtitle1">Mouse captured</Typography>
                <Typography variant="body2" color="text.secondary">
                  Move the mouse to move the cursor. Left, right and middle
                  clicks and dragging all work like a real mouse.
                </Typography>
                <Typography variant="body2" sx={{ fontWeight: 600 }}>
                  Press Esc to release
                </Typography>
              </>
            ) : (
              <Typography variant="body2" color="text.secondary">
                {padHint}
              </Typography>
            )}
          </Box>

          {canCapture && mode === "relative" && (
            <Button
              variant="outlined"
              sx={{ width: "100%" }}
              startIcon={<LockIcon />}
              disabled={captured}
              onClick={requestCapture}
            >
              {captured ? "Captured — press Esc to release" : "Capture Mouse"}
            </Button>
          )}

          {mode === "relative" && (
            <Box sx={{ px: 1 }}>
              <Typography variant="caption" color="text.secondary">
                Sensitivity
              </Typography>
              <Slider
                value={sensitivity}
                min={0.5}
                max={4}
                step={0.1}
                marks={[
                  { value: 0.5, label: "0.5×" },
                  { value: 4, label: "4×" },
                ]}
                valueLabelDisplay="auto"
                valueLabelFormat={(v) => `${v}×`}
                onChange={(_, value) => setSensitivity(value as number)}
              />
            </Box>
          )}

          <Button
            variant={dragLock ? "contained" : "outlined"}
            color={dragLock ? "warning" : "primary"}
            sx={{ width: "100%" }}
            startIcon={<PanToolIcon />}
            disabled={captured}
            onClick={toggleDragLock}
          >
            {dragLock ? "Release" : "Click & Hold"}
          </Button>

          <Grid container spacing={1}>
            {buttons.map((b) => (
              <Grid item xs={6} key={b.button}>
                <Button
                  variant="outlined"
                  sx={{ width: "100%" }}
                  disabled={captured}
                  onClick={() => clickButton(b.button)}
                >
                  {b.label}
                </Button>
              </Grid>
            ))}
          </Grid>
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Close</Button>
      </DialogActions>
    </Dialog>
  );
}
