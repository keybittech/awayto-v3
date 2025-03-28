import { SxProps, Theme as MuiTheme, createTheme } from '@mui/material';
import { red, green } from '@mui/material/colors';

export type PaletteMode = 'light' | 'dark';

declare module '@mui/material/Button' {
  interface ButtonPropsVariantOverrides {
    underline: true;
  }
}

type Theme = { theme: MuiTheme };

const drawerWidth = 175;
const paletteTheme = createTheme({
  colorSchemes: {
    light: {
      palette: {
        background: {
          default: '#ddeeff'
        },
        primary: {
          main: '#121f31',
          dark: 'rgba(0, 0, 0, .2)',
          light: '#009cc822',
          contrastText: '#333',
        },
        secondary: {
          main: '#000',
          dark: '#009cc8',
          light: '#009cc822',
          contrastText: 'rgb(190, 222, 255)'
        }
      }
    },
    dark: {
      palette: {
        background: {
          default: '#17222a'
        },
        primary: {
          main: '#ddd',
          dark: '#203040',
          light: 'rgba(125, 225, 255, 0.05)',
          contrastText: '#333',
        },
        secondary: {
          main: '#009cc8',
          dark: '#203040',
          light: 'rgba(125, 225, 255, 0.05)',
          contrastText: '#002f4a'
        }
      }
    }
  }
});


export const theme = createTheme(paletteTheme, {
  components: {
    MuiPickersDay: {
      styleOverrides: {
        root: {
          '&.Mui-selected': {
            backgroundColor: '#009cc8 !important'
          },
        }
      }
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          '& .MuiList-padding': {
            paddingLeft: 'unset'
          },
          '& .MuiListItem-button': {
            paddingLeft: '16px'
          }
        }
      }
    },
    MuiTableCell: {
      styleOverrides: {
        root: {
          padding: '4px 8px !important'
        }
      }
    },
    MuiTableBody: {
      styleOverrides: {
        root: {
          '& .MuiTableCell-body:not(:last-child)': {
            '&:not(:last-child)': {
              borderRight: '1px solid rgb(228, 228, 228)',
            }
          },
          '& .MuiIconButton-root': {
            padding: 0
          },
          '& .MuiButton-textSizeSmall': {
            padding: '0 4px'
          }
        }
      }
    },
    MuiSlider: {
      styleOverrides: {
        root: {
          padding: '4px 0'
        }
      }
    },
    MuiButton: {
      styleOverrides: {
        root: (_: Theme) => ({
          variants: [{
            props: { variant: "underline" },
            style: {
              textTransform: 'capitalize',
              textAlign: 'left',
              justifyContent: 'left',
              borderRadius: 0,
              textDecoration: 'underlined',
              background: `linear-gradient(to top, ${_.theme.palette.secondary.light} 0%, transparent 33%)`,
              borderBottom: `1px solid ${_.theme.palette.secondary.light}`,
              '&:hover': {
                background: `linear-gradient(to top, ${_.theme.palette.secondary.light} 0%, transparent 100%)`
              },
            },
          }],
          // marginBottom: '4px',
          padding: '6px 8px 4px',
          alignItems: 'baseline',
        })
      }
    },
    MuiDialog: {
      styleOverrides: {
        container: {
          '.MuiCard-root': {
            overflowY: 'scroll'
          }
        }
      }
    },
    MuiFormLabel: {
      styleOverrides: {
        asterisk: {
          color: 'red',
          fontSize: '32px',
          float: 'right'
        }
      }
    },
    MuiDataGrid: {
      styleOverrides: {
        root: (_: Theme) => ({
          borderImageSlice: 1,
          borderImageSource: `linear-gradient(to right, ${_.theme.palette.primary.dark} 50%, transparent 100%)`,
          '--DataGrid-rowBorderColor': 'none',
          '& .MuiDataGrid-row': {
            borderBottom: `1px solid ${_.theme.palette.primary.dark}`,
            borderImageSlice: 1,
            borderImageSource: `linear-gradient(to right, ${_.theme.palette.primary.dark} 50%, transparent 100%)`,
          }
        }),
        footerContainer: {
          borderTop: 'none'
        }
      }
    },
    MuiPaper: {
      styleOverrides: {
        root: (_: Theme) => ({
          variants: [{
            props: { variant: "outlined" },
            style: {
              backgroundColor: 'transparent',
              borderImageSlice: 1,
              borderImageSource: `linear-gradient(to right, ${_.theme.palette.primary.dark} 50%, transparent 100%)`,
            },
          }],
        }),
      }
    },
    MuiCardActionArea: {
      styleOverrides: {
        root: (_: Theme) => ({
          '&.actionBtnFade .MuiCardActionArea-focusHighlight': {
            background: `linear-gradient(to right, ${_.theme.palette.secondary.main} 50%, transparent 100%)`
          }
        })
      }
    }
  }
});

export const useStyles = (): { [key: string]: SxProps } => {
  const absoluteFullChild = {
    position: 'absolute',
    width: '100%',
    height: '100%'
  };
  const disabledOverrideClear = {
    '&.Mui-disabled': {
      color: '#444',
      backgroundColor: 'transparent',
    }
  };

  return {
    appLogo: {
      width: '64px'
    },
    logo: {
      width: '64px'
    },
    root: {
      display: 'flex'
    },
    siteTitle: {
      fontSize: '1.5rem',
      textAlign: 'center'
    },
    menuText: {
      fontSize: '.75rem'
    },
    colorBox: {
      width: '30px',
      height: '30px',
      display: 'block',
      margin: '12px',
      border: '1px solid #333',
      cursor: 'pointer',
      '&:hover': {
        opacity: .5
      }
    },
    appBar: {
      width: `calc(100% - ${drawerWidth}px)`,
      marginLeft: drawerWidth,
      backgroundColor: '#666'
    },
    menuIcon: {
      "&:hover svg": {
        color: 'rgb(39 109 129)'
      }
    },
    loginWrap: {
      height: '75vh'
    },
    link: {
      textDecoration: 'none'
    },
    dropzone: {
      width: '400px',
      height: '150px'
    },
    datatable: {
      borderRadius: '4px'
    },
    legendBox: {
      borderRadius: '12px',
      borderColor: 'rgba(255, 255, 255, .255)'
    },
    infoHeader: {
      fontWeight: 500,
      fontSize: '1rem',
      textTransform: 'uppercase',
      color: '#aaa !important'
    },
    infoLabel: {
      fontWeight: 500,
      fontSize: '1rem'
    },
    infoCard: {
      height: '200px',
      overflowY: 'auto'
    },
    darkRounded: {
      borderRadius: '16px',
      backgroundColor: '#203040',
      '& *': { color: 'white' },
      '&.MuiButton-root:hover': {
        backgroundColor: '#606060'
      }
    },
    green: {
      color: green[500]
    },
    red: {
      color: red[500]
    },
    onboardingProgress: {
      ...disabledOverrideClear,
      backgroundColor: 'transparent',
      color: '#fff',
      fontWeight: 700,
      fontSize: '2rem',
      width: { xs: '80px', sm: '100px', md: '120px' },
      height: '100%',
      alignItems: 'center'
    },
    audioButton: {
      cursor: 'pointer'
    },
    overflowEllipsis: {
      textOverflow: 'ellipsis',
      whiteSpace: 'nowrap',
      overflow: 'hidden'
    },
    blueChecked: {
      '& .MuiSvgIcon-root': {
        color: 'lightblue'
      }
    },
    chipRoot: {
      root: {
        margin: theme.spacing(1),
        height: '100%',
        display: 'flex',
        flexDirection: 'row'
      }
    },
    chipLabel: {
      label: {
        overflowWrap: 'break-word',
        whiteSpace: 'normal',
        textOverflow: 'clip'
      }
    },
    pdfViewerComps: {
      ...absoluteFullChild,
      display: 'flex',
      placeContent: 'center',
      placeItems: 'center',
      '& *': {
        maxWidth: '100%',
        maxHeight: '100%',
      }
    },
    whiteboardActionButton: {
      position: 'absolute',
      zIndex: 11,
      backgroundColor: '#eee',
      right: 35
    },
    variableText: {
      [theme.breakpoints.down('md')]: {
        fontSize: '24px',
      },
      [theme.breakpoints.up('md')]: {
        fontSize: '20px',
      },
    },
    variableButtonIcon: {
      [theme.breakpoints.down('sm')]: {
        fontSize: '24px',
      },
      [theme.breakpoints.up('md')]: {
        fontSize: '12px !important',
      },
    },
    mdHide: {
      [theme.breakpoints.down('md')]: {
        display: 'flex',
      },
      [theme.breakpoints.up('md')]: {
        display: 'none'
      },
    },
    mdShow: {
      [theme.breakpoints.down('md')]: {
        display: 'none',
      },
      [theme.breakpoints.up('md')]: {
        display: 'flex'
      },
    }
  }
};
