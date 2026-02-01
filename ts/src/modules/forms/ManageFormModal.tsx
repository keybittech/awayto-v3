import React, { useEffect, useMemo, useState } from 'react';

import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActions from '@mui/material/CardActions';
import Grid from '@mui/material/Grid';
import MenuItem from '@mui/material/MenuItem';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

import { siteApi, useUtil, IForm, IFormVersion, IField, deepClone, IGroupForm, targets, IFormTemplate } from 'awayto/hooks';
import FormBuilder from './FormBuilder';

interface ManageFormModalProps extends IComponent {
  editForm?: IForm;
}

export function ManageFormModal({ editForm, closeModal }: ManageFormModalProps): React.JSX.Element {
  const { setSnack } = useUtil();

  const [version, setVersion] = useState({ form: {} } as IFormVersion);
  const [form, setForm] = useState({ name: '', ...editForm } as IForm);
  const [editable, setEditable] = useState(true);
  const [groupFormId, setGroupFormId] = useState<string | null>();
  const [groupRoleIds, setGroupRoleIds] = useState<string[]>([]);

  const [postGroupFormVersion] = siteApi.useGroupFormServicePostGroupFormVersionMutation();
  const [postGroupForm] = siteApi.useGroupFormServicePostGroupFormMutation();
  const [patchGroupFormVersion] = siteApi.useGroupFormServicePatchGroupFormVersionMutation();
  const [patchGroupFormActiveVersion] = siteApi.useGroupFormServicePatchGroupFormActiveVersionMutation();
  const { data: groupRolesRequest } = siteApi.useGroupRoleServiceGetGroupRolesQuery();
  const { data: formRequest } = siteApi.useGroupFormServiceGetGroupFormByIdQuery({ formId: editForm?.id! }, { skip: !editForm });

  const versionIds = useMemo(() => {
    if (!formRequest?.groupForm.form?.versions) return [];
    return formRequest.groupForm.form.versions.map(v => v.id);
  }, [formRequest]);

  const generateForm = () => {
    setEditable(false);

    if (!form.name || !Object.keys(version.form).length || Object.values(version.form).some(v => v.some(f => !f.l))) {
      setSnack({ snackType: 'error', snackOn: 'Forms must have a name, and at least 1 field. All fields must have a label.' });
      setEditable(true);
      return;
    }

    const newForm = Object.keys(version.form).reduce((m, k, i) => {
      const fields = [...version.form[k]] as IField[];
      return {
        ...m,
        [i]: fields.map(f => {
          if ('' === f.v) delete f.v;
          if ('' === f.d) delete f.d;
          if ('' === f.h) delete f.h;
          if ('' === f.x) delete f.x;
          if (false === f.r) delete f.r;
          if (!f.o?.length) delete f.o;
          return f;
        })
      }
    }, {});

    return newForm as IFormTemplate;
  }

  useEffect(() => {
    if (!formRequest) return;
    const gf = formRequest.groupForm;
    if (gf.form) {
      setGroupFormId(gf.id);
      setGroupRoleIds(formRequest.groupRoleIds);
      setForm(gf.form as IForm);
      if (gf.form.versions) {
        const activeVersion = gf.form.versions.find(fv => fv.active);
        setVersion(activeVersion as IFormVersion);
      }
    }
  }, [formRequest]);

  useEffect(() => {
    const formClone = deepClone(form);
    if (formClone.version && Object.keys(formClone.version).length) {

      const vers = formClone.version;

      Object.keys(vers.form).forEach(k => {
        vers.form[k].forEach(f => {
          if (!f.t) f.t = 'text';
          if (!f.h) f.h = '';
          if (!f.r) f.r = false;
        });
      });

      setVersion(prev => ({
        ...vers,
        id: vers.id || prev.id
      }));
    }
  }, [form]);

  return <Card sx={{ display: 'flex', flex: 1, flexDirection: 'column' }}>
    <CardHeader
      title={`${editForm?.id ? 'Edit' : 'Create'} Form`}
      subheader="Forms must have a name, and at least 1 field. All fields must have a label. Forms can be updated up until they start receiving submissions. Forms with submissions will generate a new version upon being updated, to preserve data consistency."
    />
    <CardContent sx={{ display: 'flex', flex: 1, flexDirection: 'column', overflow: 'auto' }}>
      <Grid container spacing={2} mb={4}>
        <Grid size={6}>
          <TextField
            {...targets(`manage form modal name`, `Form Name`, `change the name of the form being edited or created`)}
            fullWidth
            autoFocus
            value={form.name}
            onChange={e => setForm({ ...form, name: e.target.value })}
          />
        </Grid>
        {groupRolesRequest?.groupRoles && groupRolesRequest?.groupRoles.length && <Grid size={6}>
          <TextField
            {...targets(`manage form modal group roles selection`, `Roles`, `select the roles which may see this form`)}
            select
            fullWidth
            helperText={'Only users with the specified roles will be able to see this form.'}
            required
            onChange={e => setGroupRoleIds(e.target.value as unknown as string[])}
            value={groupRoleIds}
            slotProps={{
              select: {
                multiple: true,
              }
            }}
          >
            {groupRolesRequest?.groupRoles.map(groupRole => {
              return <MenuItem key={`${groupRole.id}_form_role_select`} value={groupRole.id}>
                {groupRole.name}
              </MenuItem>
            })}
          </TextField>
        </Grid>}
        {!!versionIds.length && <Grid size={6}>
          <TextField
            {...targets(`form edit version selection`, `Version`, `select a form version to edit`)}
            select
            fullWidth
            value={version.id || ''}
            onChange={e => {
              if (formRequest?.groupForm.form?.versions) {
                setVersion(formRequest.groupForm.form.versions.find(v => v.id === e.target.value) as IFormVersion);
              }
            }}
          >
            {versionIds.map((vid, i) => <MenuItem key={`report_field_sel${i}`} value={vid}>Version {versionIds.length - i}</MenuItem>)}
          </TextField>
        </Grid>}
        {version.hasSubmissions && <Grid size={6}>
          <Alert color="warning" variant="outlined">
            <Typography variant="caption">
              This version has existing data submissions and cannot be directly modified. A new version will be generated instead.
            </Typography>
          </Alert>
        </Grid>}
      </Grid>

      <FormBuilder editable={editable} version={version} setVersion={setVersion} />

    </CardContent>
    <CardActions>
      <Grid size="grow" container justifyContent="space-between">
        <Button
          {...targets(`manage form modal close`, `close the form management modal`)}
          color="error"
          onClick={closeModal}
        >Cancel</Button>
        <Grid>
          {version.id && !version.active && <Button
            {...targets(`manage form modal set active version`, `set the currently selected version to be active`)}
            color="info"
            onClick={() => {
              const { id: formVersionId, formId } = version;
              patchGroupFormActiveVersion({ patchGroupFormActiveVersionRequest: { formId, formVersionId } }).unwrap().then(closeModal);
            }}
          >Set as Active</Button>}
          {!groupFormId ? <Button
            {...targets(`manage form modal submit new form`, `submit the the page to create a new form`)}
            color="info"
            onClick={() => {
              const newForm = generateForm();
              if (newForm) {
                postGroupForm({
                  postGroupFormRequest: {
                    name: form.name,
                    groupRoleIds,
                    groupForm: {
                      form: {
                        name: form.name,
                        version: {
                          form: newForm
                        }
                      }
                    } as IGroupForm
                  }
                }).unwrap().then(closeModal);
              }
            }}
          >Create Form</Button> :
            version.hasSubmissions ? <Button
              {...targets(`manage form modal submit new version`, `submit the page to create a new version of an existing form`)}
              color="info"
              onClick={() => {
                const newForm = generateForm();
                if (newForm) {
                  postGroupFormVersion({
                    postGroupFormVersionRequest: {
                      name: form.name,
                      groupFormId,
                      formId: form.id,
                      groupRoleIds,
                      groupFormVersion: {
                        form: newForm,
                        formId: form.id
                      } as IFormVersion
                    }
                  }).unwrap().then(closeModal);
                }
              }}
            >Create Version</Button> :
              <Button
                {...targets(`manage form modal update version`, `submit the page to update this version of an existing form`)}
                color="info"
                onClick={() => {
                  const newForm = generateForm();
                  if (newForm) {
                    patchGroupFormVersion({
                      patchGroupFormVersionRequest: {
                        groupFormVersion: {
                          form: newForm,
                          formId: form.id
                        }
                      }
                    }).unwrap().then(closeModal);
                  }
                }}
              >Update Version</Button>
          }
        </Grid>
      </Grid>
    </CardActions>
  </Card>
}

export default ManageFormModal;
