import { useQuery } from "@tanstack/react-query";
import { ControlApi } from "../lib/api";
import Box from "@mui/material/Box";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import { ListItemSecondaryAction } from "@mui/material";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";

function ipOrHostname(ip: string, hostname: string) {
  const regex = /^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.|$)){4}$/gm;
  if (regex.test(window.location.hostname)) {
    return ip;
  } else {
    return hostname;
  }
}

export function Network() {
  const api = new ControlApi();
  const peers = useQuery({
    queryKey: ["settings", "remote", "peers"],
    queryFn: api.getPeers,
  });

  const refreshButton = (
    <Button variant="text" onClick={() => peers.refetch()} sx={{ width: 1 }}>
      Refresh
    </Button>
  );

  if (peers.data?.peers.length === 0) {
    return (
      <Box sx={{ m: 1 }}>
        <Typography textAlign="center">No MiSTers found on network.</Typography>
      </Box>
    );
  }

  return (
    <Box sx={{ m: 1 }}>
      {refreshButton}
      <List sx={{ pt: 0, pb: 0 }}>
        {peers.data?.peers.map((peer) => (
          <ListItem key={peer.hostname}>
            <ListItemText
              primary={peer.hostname}
              secondary={peer.ip + " (v" + peer.version + ")"}
            />
            <ListItemSecondaryAction>
              <Button
                href={
                  "http://" + ipOrHostname(peer.ip, peer.hostname) + ":8182/"
                }
                target="_blank"
              >
                Connect
              </Button>
            </ListItemSecondaryAction>
          </ListItem>
        ))}
      </List>
    </Box>
  );
}
