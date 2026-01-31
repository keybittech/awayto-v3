import React, { useEffect, useMemo, useState, useCallback } from 'react';

import Grid from '@mui/material/Grid';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import MenuItem from '@mui/material/MenuItem';

import { siteApi, useUtil, IForm, IFormVersion, IField, deepClone, IGroupForm, targets } from 'awayto/hooks';
import FormBuilder from './FormBuilder';

interface ManageFormModalProps extends IComponent {
  editForm?: IForm;
}

export function ManageFormModal({ editForm, closeModal, ...props }: ManageFormModalProps): React.JSX.Element {

  // TODO add version selection
  // add help info if hasSubmissions is true
  // make save behavior switch on hasSubmissions
  // make patch version endpoint

  const { setSnack } = useUtil();

  const [version, setVersion] = useState({ form: {} } as IFormVersion);
  const [form, setForm] = useState({ name: '', ...editForm } as IForm);
  const [editable, setEditable] = useState(true);
  const [groupFormId, setGroupFormId] = useState<string | null>();
  const [groupRoleIds, setGroupRoleIds] = useState<string[]>([]);

  const [postGroupFormVersion] = siteApi.useGroupFormServicePostGroupFormVersionMutation();
  const [postGroupForm] = siteApi.useGroupFormServicePostGroupFormMutation();
  const [patchGroupFormActiveVersion] = siteApi.useGroupFormServicePatchGroupFormActiveVersionMutation();
  const { data: groupRolesRequest } = siteApi.useGroupRoleServiceGetGroupRolesQuery();
  const { data: formRequest } = siteApi.useGroupFormServiceGetGroupFormByIdQuery({ formId: editForm?.id! }, { skip: !editForm });

  const versionIds = useMemo(() => {
    if (!formRequest?.groupForm.form?.versions) return [];
    return formRequest.groupForm.form.versions.map(v => v.id);
  }, [formRequest]);

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

  const handleSubmit = useCallback(async () => {
    setEditable(false);
    const { id: formId, name } = form;

    if (!name || !Object.keys(version.form).length || Object.values(version.form).some(v => v.some(f => !f.l))) {
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

    if (groupFormId && formId) {
      await postGroupFormVersion({
        postGroupFormVersionRequest: {
          name,
          groupFormId,
          formId,
          groupRoleIds,
          groupFormVersion: {
            form: newForm,
            formId
          } as IFormVersion
        }
      }).unwrap();
    } else {
      await postGroupForm({
        postGroupFormRequest: {
          name,
          groupRoleIds,
          groupForm: {
            form: {
              name,
              formId,
              version: {
                form: newForm
              }
            }
          } as IGroupForm
        }
      }).unwrap();
    }

    if (closeModal)
      closeModal();
  }, [form, version.form, groupFormId, groupRoleIds]);

  return <Card sx={{ display: 'flex', flex: 1, flexDirection: 'column' }}>
    <CardHeader title={`${editForm?.id ? 'Edit' : 'Create'} Form`} />
    <CardContent sx={{ display: 'flex', flex: 1, flexDirection: 'column', overflow: 'auto' }}>
      <Grid container spacing={2} mb={4}>
        <Grid size={6}>
          <TextField
            {...targets(`manage form modal name`, `Form Name`, `change the name of the form being edited or created`)}
            fullWidth
            autoFocus
            value={form.name}
            onKeyDown={e => {
              if ('Enter' === e.key) {
                handleSubmit();
              }
            }}
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
        {version && !!versionIds.length && <Grid size={6}>
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
      </Grid>

      <FormBuilder {...props} editable={editable} version={version} setVersion={setVersion} />

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
          <Button
            {...targets(`manage form modal submit`, `submit the current form to be saved or edited`)}
            color="info"
            onClick={handleSubmit}
          >Submit</Button>
        </Grid>
      </Grid>
    </CardActions>
  </Card>
}

export default ManageFormModal;
