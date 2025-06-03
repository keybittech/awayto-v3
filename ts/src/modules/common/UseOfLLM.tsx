import React from 'react';

import Grid from '@mui/material/Grid';
import DialogTitle from '@mui/material/DialogTitle';
import DialogContent from '@mui/material/DialogContent';
import DialogActions from '@mui/material/DialogActions';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import IconButton from '@mui/material/IconButton';
import Box from '@mui/material/Box';

import CloseIcon from '@mui/icons-material/Close';

import { targets } from 'awayto/hooks';

import PreAiImage from '../../img/pre_ai.png';
import PostAiImage from '../../img/post_ai.png';

export function UseOfLLMModal({ closeModal }: IComponent): React.JSX.Element {
  return <>
    <DialogTitle>Use of Large-Language Models</DialogTitle>
    <IconButton
      aria-label="close"
      onClick={closeModal}
      sx={(theme) => ({
        position: 'absolute',
        right: 8,
        top: 8,
        color: theme.palette.grey[500],
      })}
    >
      <CloseIcon />
    </IconButton>
    <DialogContent>
      <Box sx={{ m: 2, '& > *': { mb: 2 } }}>
        <Typography variant="h5" component="h2" gutterBottom>
          Guiding Principle
        </Typography>
        <Typography variant="body1">
          Large-language models (LLMs) will only be used as a supplement to basic functionality on the site. There is no intention to use LLMs to automate user workflows, intercept business data for advanced use cases, or provide direct access to a chatbot or other similar technology.
        </Typography>

        <Typography variant="h5" component="h2" gutterBottom>
          Opt-In, Off by Default
        </Typography>
        <Typography variant="body1">
          An option is provided during group creation to enable AI suggestions across the group. This can be toggled on or off at any time later on, using the group configuration screens.
        </Typography>

        <Typography variant="h5" component="h2" gutterBottom>
          What models are in use?
        </Typography>
        <Typography variant="body1">
          At this time, the OpenAI API is used to facilitate LLM functionality as a means of bootstrapping faster development. In the future, the platform will utilize local language models, namely Ollama, a leading free and open-source local LLM framework.
        </Typography>

        <Typography variant="h5" component="h2" gutterBottom>
          What basic features are supplemented by LLMs?
        </Typography>
        <Typography variant="body1">
          LLMs are used to provide suggestions pertaining to the setup and configuration of a Group along with its supporting objects (Roles, Services, Tiers, and Features).
        </Typography>

        <Typography variant="h5" component="h2" gutterBottom>
          Group Naming Suggestions
        </Typography>
        <Typography variant="body1">
          Some groups or organizations which use the site may find it difficult to enumerate the concepts and naming structures behind their services, when it comes to declaring them in a schedulable format that users will understand. Therefore, AI is used to help supplement idea generation when it comes to creating a group.
        </Typography>
        <Typography variant="body1">
          Upon group creation, the group owner must provide a group name and short phrase of the group's purpose.&nbsp;
          <Typography component="strong" fontWeight="bold">
            Only group name & group purpose -- along with the names of subsequently-created Roles, Services, Tiers, or Features -- are used to generate suggestions.
          </Typography>
        </Typography>

        <Typography variant="h6" component="h3" gutterBottom>
          What does it look like?
        </Typography>
        <Typography variant="body1">
          When creating group objects (Roles, Services, Tiers, and Features) the user will be made aware that LLM is in use by the following method:
        </Typography>

        <Box component="ol" sx={{ pl: 3 }}>
          <Box component="li" sx={{ mb: 2 }}>
            <Typography variant="body1">
              Before a supported form element has fetched suggestions, the form element will show generic examples, denoted with "Ex:"
            </Typography>
            <Box sx={{ textAlign: 'center', my: 2 }}>
              <img
                src={PreAiImage}
                alt="An HTML dropdown input with example options listed under; the options are not clickable"
                style={{ maxWidth: '100%', height: 'auto' }}
              />
            </Box>
          </Box>

          <Box component="li" sx={{ mb: 2 }}>
            <Typography variant="body1">
              After suggestions have been fetched, they will be shown under the form element, denoted with "LLM:"
            </Typography>
            <Box sx={{ textAlign: 'center', my: 2 }}>
              <img
                src={PostAiImage}
                alt="An HTML dropdown input with AI-generated options listed under; each option is clickable text"
                style={{ maxWidth: '100%', height: 'auto' }}
              />
            </Box>
          </Box>

          <Box component="li" sx={{ mb: 2 }}>
            <Typography variant="body1">
              The user may click on a suggestion if desired, and it will be added to the field.
            </Typography>
          </Box>
        </Box>
      </Box>
    </DialogContent>
    <DialogActions>
      <Button
        {...targets(`use of ai modal close`, `close use of ai modal`)}
        onClick={() => closeModal && closeModal()}
      >
        Close
      </Button>
    </DialogActions>
  </>
}

export default UseOfLLMModal;
