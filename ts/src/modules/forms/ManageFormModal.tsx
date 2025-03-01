import React, { Suspense, useEffect, useState, useCallback } from 'react';

import Grid from '@mui/material/Grid';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardContent from '@mui/material/CardContent';
import CardHeader from '@mui/material/CardHeader';
import CardActions from '@mui/material/CardActions';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';

import { siteApi, useUtil, IForm, IFormVersion, IField, deepClone, IGroupForm, targets } from 'awayto/hooks';
import FormBuilder from './FormBuilder';

interface ManageFormModalProps extends IComponent {
  editForm?: IForm;
}

export function ManageFormModal({ editForm, closeModal, ...props }: ManageFormModalProps): React.JSX.Element {

  const [postGroupFormVersion] = siteApi.useGroupFormServicePostGroupFormVersionMutation();
  const [postGroupForm] = siteApi.useGroupFormServicePostGroupFormMutation();
  const [getGroupFormById] = siteApi.useLazyGroupFormServiceGetGroupFormByIdQuery();

  const { setSnack } = useUtil();

  const [version, setVersion] = useState({ form: {} } as IFormVersion);
  const [form, setForm] = useState({ name: '', ...editForm } as IForm);
  const [editable, setEditable] = useState(true);

  useEffect(() => {
    if (editForm) {
      getGroupFormById({ formId: editForm.id }).unwrap().then(res => {
        const gf = res.groupForm;
        if (gf.form) {
          setForm(gf.form as IForm);
          if (gf.form.version) {
            setVersion(gf.form.version as IFormVersion);
          }
        }

      }).catch(console.error);
    }
  }, [editForm]);

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

      setVersion(formClone.version);
    }
  }, [form]);

  const handleSubmit = useCallback(async () => {
    setEditable(false);
    const { id, name } = form;

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
          delete f.v;
          if ('' === f.t) delete f.t;
          if ('' === f.h) delete f.h;
          if ('' === f.x) delete f.x;
          if (false === f.r) delete f.r;
          return f;
        })
      }
    }, {});


    if (id) {
      await postGroupFormVersion({
        postGroupFormVersionRequest: {
          name,
          formId: id,
          groupFormVersion: {
            form: newForm,
            formId: id
          } as IFormVersion
        }
      }).unwrap();
    } else {
      await postGroupForm({
        postGroupFormRequest: {
          name,
          groupForm: {
            form: {
              name,
              formId: id,
              version: {
                form: newForm,
                submission: {}
              }
            }
          } as IGroupForm
        }
      }).unwrap();
    }

    if (closeModal)
      closeModal();
  }, [form, version.form]);

  return <Card sx={{ display: 'flex', flex: 1, flexDirection: 'column' }}>
    <CardHeader title={`${editForm?.id ? 'Edit' : 'Create'} Form`} />
    <CardContent sx={{ display: 'flex', flex: 1, flexDirection: 'column', overflow: 'auto' }}>
      <Box mt={2} />

      <Box mb={4}>
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
      </Box>

      <Suspense>
        <FormBuilder {...props} editable={editable} version={version} setVersion={setVersion} />
      </Suspense>

    </CardContent>
    <CardActions>
      <Grid size="grow" container justifyContent="space-between">
        <Button
          {...targets(`manage form modal close`, `close the form management modal`)}
          onClick={closeModal}
        >Cancel</Button>
        <Button
          {...targets(`manage form modal submit`, `submit the current form to be saved or edited`)}
          onClick={handleSubmit}
        >Submit</Button>
      </Grid>
    </CardActions>
  </Card>
}

export default ManageFormModal;
