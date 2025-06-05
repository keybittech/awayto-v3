import React, { useRef, useState } from 'react';

import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import IconButton from '@mui/material/IconButton';

import SubtitlesIcon from '@mui/icons-material/Subtitles';
import SubtitlesOffIcon from '@mui/icons-material/SubtitlesOff';

import KbtIcon from '../../img/kbt-icon_256w.png';
import { targets } from 'awayto/hooks';

function OnboardingVideo(_: IComponent): React.JSX.Element {

  const [topPos, setTopPos] = useState(0);
  const textRef = useRef<HTMLDivElement>(null);
  const [showSubtitles, setShowSubtitles] = useState(false);
  const [showSubtitlesBtn, setShowSubtitlesBtn] = useState(false);

  return <>
    <Grid container direction="column" sx={{ position: 'relative' }}>
      <div
        onMouseEnter={() => { setShowSubtitlesBtn(true) }}
        onMouseLeave={() => { setShowSubtitlesBtn(false) }}
        style={{ position: 'relative' }}
      >
        <video
          onTimeUpdate={event => {
            const target = event.target as { currentTime?: number, duration?: number };
            if (target.currentTime && target.duration && textRef.current) {
              setTopPos((target.currentTime / target.duration) * textRef.current.clientHeight);
            }
          }}
          controls
          loop
          poster={KbtIcon}
          src="/demos/onboarding.mp4"
          width="100%"
        />
        <Box sx={{ position: 'absolute', display: 'block', bottom: '64px', right: '12px' }}>
          <IconButton
            {...targets(`show subtitles`, `overlay subtitles on to the video`)}
            sx={{ visibility: showSubtitlesBtn ? 'visible' : 'hidden' }}
            onClick={() => { setShowSubtitles(!showSubtitles); }}
          >
            {!showSubtitles ? <SubtitlesIcon /> : <SubtitlesOffIcon />}
          </IconButton>
        </Box>
      </div>
      {showSubtitles && <Grid sx={{ position: 'absolute', left: '4px', right: '64px', bottom: '48px', backgroundColor: 'rgba(0,0,0,.1)' }}>
        <Grid container sx={{
          backgroundColor: 'rgba(0,0,0,.5)',
          overflow: 'hidden',
          height: '256px',
          position: 'relative',
          maskImage: 'linear-gradient(to bottom, transparent 5%, #000 25%, #000 85%, transparent 95%)'
        }}>
          <Grid
            size="grow"
            sx={{
              position: 'absolute',
              zIndex: 90,
              transition: 'top 0.5s ease',
              top: `-${topPos}px`,
              lineHeight: 2,
              fontWeight: 700,
              padding: '20px',
            }}
            ref={textRef}
          >
            <p>&nbsp;</p>
            <p>&nbsp;</p>
            <p>Start by providing a unique name for your group. Group name can be changed later.</p>
            <p>If AI Suggestions are enabled, the group name and description will be used to generate custom suggestions for naming roles, services, and other elements on the site.</p>
            <p>Restrict who can join your group by adding an email to the list of allowed domains. For example, site.com is the domain for the email user@site.com. To ensure only these email accounts can join the group, enter site.com into the Allowed Email Domains and press Add. Multiple domains can be added. Leave empty to allow users with any email address.</p>
            <p>To make onboarding easier, we'll use the example of creating an online learning center. For this step, we give our group a name and description which reflect the group's purpose.</p>

            <p>Roles allow access to different functionality on the site. Each user is assigned 1 role. You have the Admin role.</p>
            <p>If AI is enabled, some role name suggestions have been provided based on your group details. You can add them to your group by clicking on them. Otherwise, click the dropdown to add your own roles.</p>
            <p>Once you've created some roles, set the default role as necessary. This role will automatically be assigned to new users who join your group. Normally you would choose the role which you plan to have the least amount of access.</p>
            <p>For example, our learning center might have Student and Tutor roles. By default, everyone that joins is a Student. If a Tutor joins the group, the Admin can manually change their role in the user list.</p>

            <p>Services define the context of the appointments that happen within your group. You can add forms and tiers to distinguish the details of your service.</p>
            <p>Forms can be used to enhance the information collected before and after an appointment. Click on a form dropdown to add a new form.</p>
            <p>Each service should have at least 1 tier. The concept of the tiers is up to you, but it essentially allows for the distinction of level of service.</p>
            <p>For example, our learning center creates a service called Writing Tutoring, which has a single tier, General. The General tier has a few features: Feedback, Grammar Help, Brainstorming. Forms are used to get details about the student's paper and then ask how the appointment went afterwards.</p>

            <p>Next we create a group schedule. Start by providing basic details about the schedule and when it should be active.</p>
            <p>Some premade options are available to select common defaults. Try selecting a default and adjusting it to your exact specifications.</p>
            <p>For example, at our learning center, students and tutors meet in 30 minute sessions. Tutors work on an hours per week basis. So we create a schedule with a week duration, an hour bracket duration, and a booking slot of 30 minutes.</p>

            <h4>Review</h4>
            <p>Make sure everything looks good, then create your group.</p>
          </Grid>
        </Grid>
      </Grid>}
    </Grid>
  </>;
}

export default OnboardingVideo;
