import React, { useContext } from 'react';

import Box from '@mui/material/Box';
import Divider from '@mui/material/Divider';
import Menu from '@mui/material/Menu';
import Checkbox from '@mui/material/Checkbox';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import ListItemButton from '@mui/material/ListItemButton';
import Tooltip from '@mui/material/Tooltip';
import IconButton from '@mui/material/IconButton';

import ApprovalIcon from '@mui/icons-material/Approval';
import DoNotDisturbIcon from '@mui/icons-material/DoNotDisturb';

import { bookingFormat, targets } from 'awayto/hooks';

import PendingQuotesContext, { PendingQuotesContextType } from './PendingQuotesContext';

interface PendingQuotesMenuProps extends IComponent {
  pendingQuotesAnchorEl: null | HTMLElement;
  pendingQuotesMenuId: string;
  isPendingQuotesOpen: boolean;
  handleMenuClose: () => void;
}

export function PendingQuotesMenu({ handleMenuClose, pendingQuotesAnchorEl, pendingQuotesMenuId, isPendingQuotesOpen }: PendingQuotesMenuProps): React.JSX.Element {

  const {
    pendingQuotes,
    selectedPendingQuotes,
    handleSelectPendingQuote,
    handleSelectPendingQuoteAll,
    approvePendingQuotes,
    denyPendingQuotes
  } = useContext(PendingQuotesContext) as PendingQuotesContextType;

  return <Menu
    anchorEl={pendingQuotesAnchorEl}
    anchorOrigin={{
      vertical: 'bottom',
      horizontal: 'right',
    }}
    id={pendingQuotesMenuId}
    keepMounted
    transformOrigin={{
      vertical: 'top',
      horizontal: 'right',
    }}
    open={!!isPendingQuotesOpen}
    onClose={handleMenuClose}
  >
    <List>
      {pendingQuotes.length ? <Box sx={{ width: 300 }}>
        <ListItem
          disablePadding
          secondaryAction={!!selectedPendingQuotes.length && <>
            <Tooltip title="Approve">
              <IconButton
                {...targets(`pending requests menu approve`, `approve selected pending requests`)}
                edge="end"
                onClick={approvePendingQuotes}
              >
                <ApprovalIcon />
              </IconButton>
            </Tooltip>
            &nbsp;&nbsp;&nbsp;
            <Tooltip title="Deny">
              <IconButton
                {...targets(`pending requests menu deny`, `deny selected pending requests`)}
                edge="end"
                onClick={denyPendingQuotes}
              >
                <DoNotDisturbIcon />
              </IconButton>
            </Tooltip>
          </>}
        >
          <ListItemButton role={undefined} dense>
            <ListItemIcon>
              <Checkbox
                {...targets(`pending requests menu select all`, `select all pending requests in the list`)}
                disableRipple
                tabIndex={-1}
                onClick={handleSelectPendingQuoteAll}
                checked={selectedPendingQuotes.length === pendingQuotes.length && pendingQuotes.length !== 0}
                indeterminate={selectedPendingQuotes.length !== pendingQuotes.length && selectedPendingQuotes.length !== 0}
                disabled={pendingQuotes.length === 0}
              />
            </ListItemIcon>
            <ListItemText primary="Pending Requests" />
          </ListItemButton>
        </ListItem>

        <Divider />

        {pendingQuotes.map((pq, i) => {
          return <ListItem
            key={`pending_quotes_pqs_${i}`}
            disablePadding
          >
            <ListItemButton role={undefined} onClick={() => handleSelectPendingQuote(pq.id)} dense>
              <ListItemIcon>
                <Checkbox
                  {...targets(`pending requests menu select ${i}`, `select a single pending request from the list`)}
                  checked={selectedPendingQuotes.indexOf(pq.id) !== -1}
                  tabIndex={-1}
                  disableRipple
                />
              </ListItemIcon>
              <ListItemText
                primary={`${bookingFormat(pq.slotDate, pq.startTime)}`}
                secondary={`${pq.serviceName} ${pq.serviceTierName}`}
              />
            </ListItemButton>
          </ListItem>
        })}
      </Box> : <Box sx={{ width: 250 }}>
        <ListItem>
          <ListItemText>No pending requests.</ListItemText>
        </ListItem>
      </Box>}
    </List>
  </Menu>
}

export default PendingQuotesMenu;
