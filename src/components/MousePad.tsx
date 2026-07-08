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

import { ControlApi } from "../lib/api";
import { MouseButtons, ScreenResponse } from "../lib/models";

type MouseMode = "relative" | "absolute";

// movement (in CSS pixels) below which a press/release is treated as a tap
const tapThreshold = 10;
// maximum duration (ms) of a tap
const tapDuration = 300;

interface MousePadProps {
  open: boolean;
  onClose: () => void;
  sendMessage: (message: string) => void;
}

export default function MousePad(props: MousePadProps) {
  const { open, onClose, sendMessage } = props;

  const [mode, setMode] = useState<MouseMode>("relative");
  const [sensitivity, setSensitivity] = useState<number>(1.5);
  const [screen, setScreen] = useState<ScreenResponse | null>(null);
  const [screenError, setScreenError] = useState<boolean>(false);

  const padRef = useRef<HTMLDivElement | null>(null);
  const lastPointRef = useRef<{ x: number; y: number } | null>(null);
  const movedRef = useRef<number>(0);
  const downTimeRef = useRef<number>(0);
  const pendingRef = useRef<{ dx: number; dy: number }>({ dx: 0, dy: 0 });
  const rafRef = useRef<number | null>(null);

  // keep the latest mode/sensitivity readable from pointer handlers
  const modeRef = useRef<MouseMode>(mode);
  const sensitivityRef = useRef<number>(sensitivity);
  useEffect(() => {
    modeRef.current = mode;
  }, [mode]);
  useEffect(() => {
    sensitivityRef.current = sensitivity;
  }, [sensitivity]);

  // the screen resolution is needed to map the pad onto absolute coordinates
  useEffect(() => {
    if (!open) {
      return;
    }

    let cancelled = false;
    const api = new ControlApi();
    api
      .getScreen()
      .then((res) => {
        if (!cancelled) {
          setScreen(res);
          setScreenError(false);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setScreen(null);
          setScreenError(true);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [open]);

  // cancel any queued movement frame when the dialog closes
  useEffect(() => {
    if (!open && rafRef.current !== null) {
      window.cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
      pendingRef.current = { dx: 0, dy: 0 };
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
    if (!pad || !screen) {
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

    const x = Math.round(fx * (screen.width - 1));
    const y = Math.round(fy * (screen.height - 1));
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

    // a quick tap in trackpad mode acts as a left click
    if (wasTap && modeRef.current === "relative") {
      sendMessage("mouseBtn:click");
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
  if (mode === "absolute") {
    if (screenError) {
      padHint = "Screen resolution unavailable";
    } else if (screen) {
      padHint = `Tap or drag to position · ${screen.width}×${screen.height}`;
    } else {
      padHint = "Reading screen resolution…";
    }
  } else {
    padHint = "Drag to move · tap to click";
  }

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
