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

interface MousePadProps {
  open: boolean;
  onClose: () => void;
  sendMessage: (message: string) => void;
}

export default function MousePad(props: MousePadProps) {
  const { open, onClose, sendMessage } = props;

  const [mode, setMode] = useState<MouseMode>("relative");
  const [sensitivity, setSensitivity] = useState<number>(1.5);

  const padRef = useRef<HTMLDivElement | null>(null);
  const lastPointRef = useRef<{ x: number; y: number } | null>(null);
  const movedRef = useRef<number>(0);
  const downTimeRef = useRef<number>(0);
  const pendingRef = useRef<{ dx: number; dy: number }>({ dx: 0, dy: 0 });
  const rafRef = useRef<number | null>(null);
  const lastTapTimeRef = useRef<number | null>(null);
  const lastTapPosRef = useRef<{ x: number; y: number } | null>(null);
  const clickTimerRef = useRef<number | null>(null);

  // keep the latest mode/sensitivity readable from pointer handlers
  const modeRef = useRef<MouseMode>(mode);
  const sensitivityRef = useRef<number>(sensitivity);
  useEffect(() => {
    modeRef.current = mode;
  }, [mode]);
  useEffect(() => {
    sensitivityRef.current = sensitivity;
  }, [sensitivity]);

  // cancel any queued movement frame and pending tap when the dialog closes
  useEffect(() => {
    if (!open) {
      if (rafRef.current !== null) {
        window.cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
        pendingRef.current = { dx: 0, dy: 0 };
      }
      if (clickTimerRef.current !== null) {
        window.clearTimeout(clickTimerRef.current);
        clickTimerRef.current = null;
      }
      lastTapTimeRef.current = null;
      lastTapPosRef.current = null;
    }
  }, [open]);

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

  const handlePointerDown = (e: React.PointerEvent<HTMLDivElement>) => {
    e.currentTarget.setPointerCapture(e.pointerId);
    lastPointRef.current = { x: e.clientX, y: e.clientY };
    movedRef.current = 0;
    downTimeRef.current = e.timeStamp;

    if (modeRef.current === "absolute") {
      sendAbsolute(e.clientX, e.clientY);
    }
  };

  const handlePointerMove = (e: React.PointerEvent<HTMLDivElement>) => {
    const last = lastPointRef.current;
    if (!last) {
      return;
    }

    if (modeRef.current === "absolute") {
      sendAbsolute(e.clientX, e.clientY);
      lastPointRef.current = { x: e.clientX, y: e.clientY };
      return;
    }

    const rawDx = e.clientX - last.x;
    const rawDy = e.clientY - last.y;
    movedRef.current += Math.abs(rawDx) + Math.abs(rawDy);
    pendingRef.current.dx += rawDx * sensitivityRef.current;
    pendingRef.current.dy += rawDy * sensitivityRef.current;
    lastPointRef.current = { x: e.clientX, y: e.clientY };
    scheduleFlush();
  };

  const handlePointerEnd = (e: React.PointerEvent<HTMLDivElement>) => {
    if (lastPointRef.current === null) {
      return;
    }

    const duration = e.timeStamp - downTimeRef.current;
    const wasTap = movedRef.current < tapThreshold && duration < tapDuration;
    lastPointRef.current = null;
    if (!wasTap) {
      return;
    }

    // a quick tap acts as a left click; two quick taps as a double click,
    // in both trackpad and absolute modes
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

  const clickButton = (button: MouseButtons) => {
    sendMessage(`mouseBtn:${button}`);
  };

  const buttons: { label: string; button: MouseButtons }[] = [
    { label: "Left Click", button: "click" },
    { label: "Double Click", button: "double_click" },
    { label: "Right Click", button: "right" },
    { label: "Middle Click", button: "middle" },
  ];

  const padHint =
    mode === "absolute"
      ? "Drag to position · tap/double-tap to click"
      : "Drag to move · tap to click · double-tap to double-click";

  return (
    <Dialog open={open} onClose={onClose} fullWidth>
      <DialogContent>
        <Stack spacing={2}>
          <ToggleButtonGroup
            color="primary"
            value={mode}
            exclusive
            fullWidth
            size="small"
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
            onPointerDown={handlePointerDown}
            onPointerMove={handlePointerMove}
            onPointerUp={handlePointerEnd}
            onPointerCancel={handlePointerEnd}
            sx={{
              height: "40vh",
              minHeight: 200,
              border: 1,
              borderColor: "divider",
              borderRadius: 2,
              bgcolor: "action.hover",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              textAlign: "center",
              px: 2,
              touchAction: "none",
              userSelect: "none",
              cursor: mode === "absolute" ? "crosshair" : "move",
            }}
          >
            <Typography variant="body2" color="text.secondary">
              {padHint}
            </Typography>
          </Box>

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

          <Grid container spacing={1}>
            {buttons.map((b) => (
              <Grid item xs={6} key={b.button}>
                <Button
                  variant="outlined"
                  sx={{ width: "100%" }}
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
