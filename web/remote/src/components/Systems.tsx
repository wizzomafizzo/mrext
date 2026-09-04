import React from "react";
import { useQuery, useMutation } from "@tanstack/react-query";

import Button from "@mui/material/Button";

import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";

import { BottomNavigation, BottomNavigationAction } from "@mui/material";

import PlayArrowIcon from "@mui/icons-material/PlayArrow";

import { ControlApi } from "../lib/api";
import { System } from "../lib/models";

export default function Systems() {
  const api = new ControlApi();

  const allSystems = useQuery({
    queryKey: ["systems"],
    queryFn: api.getSystems,
  });

  const launchCore = useMutation({
    mutationFn: api.launchSystem,
  });

  const [tab, setTab] = React.useState(0);

  const getCategory = (category: string) => {
    return allSystems.data
      ?.filter((system) => system.category === category)
      .sort((a, b) => a.name.localeCompare(b.name));
  };

  const launchEntry = (system: System) => {
    return (
      <ListItem
        key={system.id}
        secondaryAction={
          <Button
            variant="text"
            onClick={() => launchCore.mutate(system.id)}
            startIcon={<PlayArrowIcon />}
          >
            Launch
          </Button>
        }
      >
        <ListItemText primary={system.name} />
      </ListItem>
    );
  };

  return (
    <div>
      {/* <div style={{ maxWidth: "100%" }}>
                <Tabs
                    value={tab}
                    onChange={(event, newValue) => {
                        setTab(newValue);
                    }}
                    textColor="secondary"
                    indicatorColor="secondary"
                    variant="scrollable"
                    scrollButtons="auto"
                >
                    <Tab label="Consoles" />
                    <Tab label="Computers" />
                    <Tab label="Other" />
                    <Tab label="Handhelds" />
                </Tabs>
            </div> */}
      <List sx={{ mb: "56px" }}>
        {tab === 0
          ? getCategory("Console")?.map((system) => launchEntry(system))
          : null}
        {tab === 1
          ? getCategory("Computer")?.map((system) => launchEntry(system))
          : null}
        {tab === 2
          ? getCategory("Other")?.map((system) => launchEntry(system))
          : null}
      </List>
      <BottomNavigation
        sx={{
          position: "fixed",
          bottom: 0,
          left: 0,
          right: 0,
          ml: { sm: "240px" },
        }}
        showLabels
        value={tab}
        onChange={(event, newValue) => {
          setTab(newValue);
        }}
      >
        <BottomNavigationAction label="Consoles" />
        <BottomNavigationAction label="Computers" />
        <BottomNavigationAction label="Other" />
      </BottomNavigation>
    </div>
  );
}
