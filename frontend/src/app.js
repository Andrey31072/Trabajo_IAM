const API_BASE = '/api/v1';
const TRAINING_CENTER_ID = '11111111-1111-1111-1111-111111111111';

const state = {
  route: location.hash.replace('#', '') || '/login',
  auth: JSON.parse(localStorage.getItem('iam_auth') || 'null'),
  me: null,
  users: [],
  roles: [],
  modules: [],
  audit: [],
  pagination: null,
  selectedUser: null,
  sessions: [],
  overrides: [],
  message: '',
  error: '',
  tab: 'roles',
  loading: false
};

const app = document.querySelector('#app');

window.addEventListener('hashchange', () => {
  state.route = location.hash.replace('#', '') || '/login';
  state.error = '';
  state.message = '';
  render();
});

document.addEventListener('submit', async (event) => {
  const form = event.target;
  if (!form.matches('[data-action]')) return;
  event.preventDefault();
  const action = form.dataset.action;
  const values = Object.fromEntries(new FormData(form).entries());
  await actions[action]?.(values, form);
});

document.addEventListener('click', async (event) => {
  const button = event.target.closest('[data-click]');
  if (!button) return;
  const name = button.dataset.click;
  await clicks[name]?.(button);
});

const actions = {
  login: async (values) => {
    await run(async () => {
      const auth = await api('/auth/login', {
        method: 'POST',
        body: values,
        skipAuth: true
      });
      state.auth = auth;
      localStorage.setItem('iam_auth', JSON.stringify(auth));
      state.me = null;
      navigate('/dashboard');
    });
  },
  forgot: async (values) => {
    await run(async () => {
      await api('/auth/password-reset/request', {
        method: 'POST',
        body: values,
        skipAuth: true,
        empty: true
      });
      state.message = 'Si el correo existe, enviaremos un enlace para restablecer la contraseña.';
      render();
    });
  },
  resetPassword: async (values, form) => {
    if (values.new_password !== values.confirm_password) {
      state.error = 'Las contraseñas no coinciden.';
      state.message = '';
      render();
      return;
    }
    await run(async () => {
      await api('/auth/password-reset/confirm', {
        method: 'POST',
        body: {
          token: values.token,
          new_password: values.new_password
        },
        skipAuth: true,
        empty: true
      });
      form.reset();
      state.message = 'Contraseña actualizada. Ya puedes iniciar sesión.';
    });
  },
  createUser: async (values, form) => {
    await run(async () => {
      const payload = {
        email: values.email,
        first_name: values.first_name,
        last_name: values.last_name,
        actor_type: values.actor_type,
        actor_id: values.actor_id || null,
        initial_role: values.initial_role || null,
        training_center_id: values.training_center_id || null
      };
      const created = await api('/users', { method: 'POST', body: payload });
      state.message = `Usuario creado. Contraseña temporal: ${created.temporary_password}`;
      form.reset();
      closeModal();
      await loadUsers();
    });
  },
  assignRole: async (values) => {
    await run(async () => {
      await api(`/users/${state.selectedUser.id}/roles`, {
        method: 'POST',
        body: {
          role_name: values.role_name,
          training_center_id: values.training_center_id || null,
          expires_at: values.expires_at || null
        }
      });
      closeModal();
      await selectUser(state.selectedUser.id);
      await loadUsers();
    });
  },
  updateUser: async (values) => {
    await run(async () => {
      await api(`/users/${state.selectedUser.id}`, {
        method: 'PUT',
        body: {
          first_name: values.first_name,
          last_name: values.last_name,
          actor_type: values.actor_type,
          actor_id: values.actor_id || null,
          is_active: values.is_active === 'true'
        }
      });
      closeModal();
      await selectUser(state.selectedUser.id);
      await loadUsers();
    });
  },
  createOverride: async (values) => {
    await run(async () => {
      await api(`/users/${state.selectedUser.id}/scope-overrides`, {
        method: 'POST',
        body: {
          feature_code: values.feature_code,
          scope_type: values.scope_type,
          is_allowed: values.is_allowed === 'true',
          reason: values.reason,
          expires_at: values.expires_at || null
        }
      });
      closeModal();
      await selectUser(state.selectedUser.id);
    });
  },
  auditFilters: async (values) => {
    await loadAudit(values);
  },
  userFilters: async (values) => {
    await loadUsers(values);
  }
};

const clicks = {
  logout: async () => {
    try {
      await api('/auth/logout', {
        method: 'POST',
        body: { refresh_token: state.auth?.refresh_token },
        empty: true
      });
    } catch (_error) {
      // Local logout still wins if the network is gone.
    }
    localStorage.removeItem('iam_auth');
    Object.assign(state, { auth: null, me: null, users: [], roles: [], modules: [], audit: [] });
    navigate('/login');
  },
  openCreateUser: () => openModal(createUserModal()),
  openAssignRole: () => openModal(assignRoleModal()),
  openEditUser: () => openModal(editUserModal()),
  openOverride: () => openModal(scopeOverrideModal()),
  closeModal,
  refreshUsers: async () => loadUsers(),
  refreshAudit: async () => loadAudit(),
  rolesTab: async (button) => {
    state.tab = button.dataset.tab;
    render();
  },
  selectUser: async (button) => selectUser(button.dataset.id),
  deactivateUser: async () => {
    if (!state.selectedUser) return;
    await run(async () => {
      await api(`/users/${state.selectedUser.id}/deactivate`, { method: 'POST', empty: true });
      await selectUser(state.selectedUser.id);
      await loadUsers();
    });
  },
  revokeSession: async (button) => {
    await run(async () => {
      await api(`/users/${state.selectedUser.id}/sessions/${button.dataset.id}`, {
        method: 'DELETE',
        empty: true
      });
      await selectUser(state.selectedUser.id);
    });
  },
  revokeRole: async (button) => {
    await run(async () => {
      await api(`/users/${state.selectedUser.id}/roles/${button.dataset.role}`, {
        method: 'DELETE',
        empty: true
      });
      await selectUser(state.selectedUser.id);
      await loadUsers();
    });
  },
  deleteOverride: async (button) => {
    await run(async () => {
      await api(`/users/${state.selectedUser.id}/scope-overrides/${button.dataset.id}`, {
        method: 'DELETE',
        empty: true
      });
      await selectUser(state.selectedUser.id);
    });
  },
  nextAudit: async (button) => {
    await loadAudit({ cursor: button.dataset.cursor });
  }
};

async function render() {
  if (!state.auth && !isPublicRoute(state.route)) {
    renderLogin();
    return;
  }

  if (state.auth && state.route === '/login') {
    navigate('/dashboard');
    return;
  }

  if (state.auth && !state.me) {
    await run(async () => {
      state.me = await api('/auth/me');
    }, false);
  }

  if (state.route.startsWith('/forgot-password')) {
    renderForgot();
    return;
  }

  if (state.route.startsWith('/reset-password')) {
    renderResetPassword();
    return;
  }

  if (state.route.startsWith('/dashboard')) {
    await ensureDashboardData();
    renderShell('dashboard');
    return;
  }

  if (state.route.startsWith('/users')) {
    await loadUsersOnce();
    renderShell('users');
    return;
  }

  if (state.route.startsWith('/rbac')) {
    await loadRbacOnce();
    renderShell('rbac');
    return;
  }

  if (state.route.startsWith('/audit')) {
    await loadAuditOnce();
    renderShell('audit');
    return;
  }

  navigate('/dashboard');
}

function renderLogin() {
  app.innerHTML = `
    <main class="auth-layout">
      <section class="auth-panel">
        ${brand()}
        <div class="auth-card">
          <h1>Ingresar</h1>
          <p>Accede con tu cuenta institucional.</p>
          ${messageBlock()}
          <form data-action="login">
            <div class="field">
              <label for="email">Correo</label>
              <input id="email" name="email" type="email" autocomplete="username" value="admin@sena.edu.co" required>
            </div>
            <div class="field">
              <label for="password">Contraseña</label>
              <input id="password" name="password" type="password" autocomplete="current-password" value="Demo2026!" required>
            </div>
            <div class="auth-actions">
              <button class="button primary" type="submit">Ingresar</button>
              <a class="button ghost" href="#/forgot-password">Recuperar contraseña</a>
            </div>
          </form>
        </div>
      </section>
      <section class="auth-visual">
        <div class="auth-visual-inner">
          <h2>IAM y Seguridad</h2>
          <p>Gestión de identidad, roles, permisos por feature y sesiones activas para la plataforma de horarios SENA.</p>
          <div class="auth-highlights">
            <span>RBAC</span>
            <span>JWT</span>
            <span>Auditoría</span>
          </div>
        </div>
      </section>
    </main>`;
}

function renderForgot() {
  app.innerHTML = `
    <main class="auth-layout">
      <section class="auth-panel">
        ${brand()}
        <div class="auth-card">
          <h1>Recuperar contraseña</h1>
          <p>Ingresa el correo institucional. La respuesta no revela si la cuenta existe.</p>
          ${messageBlock()}
          <form data-action="forgot">
            <div class="field">
              <label>Correo</label>
              <input name="email" type="email" placeholder="usuario@sena.edu.co" required>
            </div>
            <div class="auth-actions">
              <button class="button primary" type="submit">Enviar enlace</button>
              <a class="button ghost" href="#/login">Volver</a>
            </div>
          </form>
        </div>
      </section>
      <section class="auth-visual"><div class="auth-visual-inner"><h2>Recuperación segura</h2><p>Tokens temporales, expiración en una hora y revocación de sesiones al cambiar la contraseña.</p><div class="auth-highlights"><span>1 hora</span><span>Token único</span><span>Sin fuga de datos</span></div></div></section>
    </main>`;
}

function renderResetPassword() {
  const token = getRouteParam('token');
  app.innerHTML = `
    <main class="auth-layout">
      <section class="auth-panel">
        ${brand()}
        <div class="auth-card">
          <h1>Nueva contraseña</h1>
          <p>Crea una contraseña nueva usando el enlace recibido por correo.</p>
          ${messageBlock()}
          <form data-action="resetPassword">
            ${token ? `<input name="token" type="hidden" value="${escapeHtml(token)}">` : `
              <div class="field">
                <label>Token</label>
                <input name="token" type="text" required>
              </div>`}
            <div class="field">
                <label>Nueva contraseña</label>
              <input name="new_password" type="password" minlength="8" autocomplete="new-password" required>
            </div>
            <div class="field">
              <label>Confirmar contraseña</label>
              <input name="confirm_password" type="password" minlength="8" autocomplete="new-password" required>
            </div>
            <div class="auth-actions">
              <button class="button primary" type="submit">Actualizar</button>
              <a class="button ghost" href="#/login">Ingresar</a>
            </div>
          </form>
        </div>
      </section>
      <section class="auth-visual"><div class="auth-visual-inner"><h2>Cuenta protegida</h2><p>Al actualizar la contraseña se cierran las sesiones activas y el token queda marcado como usado.</p><div class="auth-highlights"><span>Sesiones cerradas</span><span>Hash bcrypt</span><span>Uso único</span></div></div></section>
    </main>`;
}

function renderShell(view) {
  app.innerHTML = `
    <main class="app-shell">
      <aside class="sidebar">
        ${brand()}
        <nav class="nav">
          ${navButton('/dashboard', 'Panel', view === 'dashboard')}
          ${navButton('/users', 'Usuarios', view === 'users')}
          ${navButton('/rbac', 'RBAC', view === 'rbac')}
          ${navButton('/audit', 'Auditoría login', view === 'audit')}
        </nav>
      </aside>
      <section class="main">
        ${topbar()}
        ${messageBlock()}
        ${view === 'dashboard' ? dashboardView() : ''}
        ${view === 'users' ? usersView() : ''}
        ${view === 'rbac' ? rbacView() : ''}
        ${view === 'audit' ? auditView() : ''}
      </section>
    </main>`;
}

function dashboardView() {
  const activeUsers = state.users.filter((user) => user.is_active).length;
  const featureCount = state.modules.reduce((total, module) => total + module.features.length, 0);
  return `
    <section class="grid metrics">
      ${metric('Usuarios activos', activeUsers)}
      ${metric('Roles de sistema', state.roles.length)}
      ${metric('Modulos', state.modules.length)}
      ${metric('Features', featureCount)}
    </section>
    <section class="grid split" style="margin-top:16px">
      <div class="panel">
        <div class="panel-header"><div><h2 class="panel-title">Sesion actual</h2><p class="panel-subtitle">${state.me?.email || ''}</p></div>${badge('green', 'Activa')}</div>
        <div class="panel-body">
          <p><strong>${escapeHtml(state.me?.full_name || '')}</strong></p>
          <p class="muted">${(state.me?.roles || []).map((r) => r.name).join(', ')}</p>
          <div class="feature-list">${(state.me?.modules || []).map((m) => badge('blue', m)).join('')}</div>
        </div>
      </div>
      <div class="panel">
        <div class="panel-header"><div><h2 class="panel-title">Credenciales demo</h2><p class="panel-subtitle">Todas usan Demo2026!</p></div></div>
        <div class="panel-body">
          <div class="feature-list">
            ${['admin@sena.edu.co', 'director@sena.edu.co', 'coordinador@sena.edu.co', 'instructor@sena.edu.co', 'aprendiz@sena.edu.co', 'administrativo@sena.edu.co'].map((email) => badge('gray', email)).join('')}
          </div>
        </div>
      </div>
    </section>`;
}

function usersView() {
  return `
    <div class="topbar">
      <div><h1>Usuarios</h1><p>Gestión de cuentas, roles, sesiones y overrides.</p></div>
      <button class="button primary" data-click="openCreateUser">Nuevo usuario</button>
    </div>
    <form class="toolbar" data-action="userFilters">
      <div class="field"><label>Buscar</label><input name="search" placeholder="Nombre o correo"></div>
      <div class="field"><label>Rol</label><select name="role"><option value="">Todos</option>${state.roles.map((r) => `<option>${r.name}</option>`).join('')}</select></div>
      <div class="field"><label>Estado</label><select name="is_active"><option value="">Todos</option><option value="true">Activo</option><option value="false">Inactivo</option></select></div>
      <button class="button blue" type="submit">Filtrar</button>
      <button class="button" type="button" data-click="refreshUsers">Actualizar</button>
    </form>
    <section class="split">
      <div class="panel">
        <div class="panel-header"><div><h2 class="panel-title">Listado</h2><p class="panel-subtitle">${state.pagination?.total_items || 0} registros</p></div></div>
        ${usersTable()}
      </div>
      <div class="panel">
        ${userDetail()}
      </div>
    </section>`;
}

function usersTable() {
  if (!state.users.length) return '<div class="empty">No hay usuarios para mostrar.</div>';
  return `
    <div class="table-wrap">
      <table>
        <thead><tr><th>Nombre</th><th>Correo</th><th>Rol</th><th>Estado</th><th></th></tr></thead>
        <tbody>
          ${state.users.map((user) => `
            <tr>
              <td><strong>${escapeHtml(user.full_name)}</strong><div class="muted">${user.actor_type}</div></td>
              <td>${escapeHtml(user.email)}</td>
              <td>${(user.roles || []).slice(0, 2).map((role) => badge('blue', role.name)).join(' ') || badge('gray', 'Sin rol')}</td>
              <td>${badge(user.is_active ? 'green' : 'red', user.is_active ? 'Activo' : 'Inactivo')}</td>
              <td><button class="button" data-click="selectUser" data-id="${user.id}">Ver</button></td>
            </tr>`).join('')}
        </tbody>
      </table>
    </div>`;
}

function userDetail() {
  const user = state.selectedUser;
  if (!user) return '<div class="panel-body empty">Selecciona un usuario para ver roles, sesiones y overrides.</div>';
  return `
    <div class="panel-header">
      <div><h2 class="panel-title">${escapeHtml(user.full_name)}</h2><p class="panel-subtitle">${escapeHtml(user.email)}</p></div>
      ${badge(user.is_active ? 'green' : 'red', user.is_active ? 'Activo' : 'Inactivo')}
    </div>
    <div class="panel-body">
      <p><strong>Tipo:</strong> ${user.actor_type}</p>
      <p><strong>Último acceso:</strong> ${formatDate(user.last_login_at)}</p>
      <div class="toolbar">
        <button class="button" data-click="openEditUser">Editar</button>
        <button class="button blue" data-click="openAssignRole">Asignar rol</button>
        <button class="button" data-click="openOverride">Override</button>
        <button class="button danger" data-click="deactivateUser">Desactivar</button>
      </div>
      <h3>Roles</h3>
      <div class="feature-list">${(user.roles || []).map((role) => `${badge('blue', role.name)} <button class="button icon danger" data-click="revokeRole" data-role="${role.name}" title="Revocar">x</button>`).join('') || badge('gray', 'Sin roles')}</div>
      <h3>Sesiones activas</h3>
      ${sessionsTable()}
      <h3>Overrides</h3>
      ${overridesTable()}
    </div>`;
}

function sessionsTable() {
  if (!state.sessions.length) return '<p class="muted">No hay sesiones activas.</p>';
  return `
    <div class="table-wrap">
      <table>
        <thead><tr><th>Dispositivo</th><th>IP</th><th>Creada</th><th>Expira</th><th></th></tr></thead>
        <tbody>
          ${state.sessions.map((session) => `
            <tr>
              <td>${escapeHtml(session.device_hint || 'Desconocido')}</td>
              <td>${escapeHtml(session.ip_address || '-')}</td>
              <td>${formatDate(session.created_at)}</td>
              <td>${formatDate(session.expires_at)}</td>
              <td><button class="button danger" data-click="revokeSession" data-id="${session.id}">Revocar</button></td>
            </tr>`).join('')}
        </tbody>
      </table>
    </div>`;
}

function overridesTable() {
  if (!state.overrides.length) return '<p class="muted">Sin overrides activos.</p>';
  return `
    <div class="table-wrap">
      <table>
        <thead><tr><th>Feature</th><th>Scope</th><th>Tipo</th><th>Expira</th><th></th></tr></thead>
        <tbody>
          ${state.overrides.map((item) => `
            <tr>
              <td class="code">${item.feature_code}</td>
              <td>${badge('gray', item.scope_type)}</td>
              <td>${badge(item.is_allowed ? 'green' : 'red', item.is_allowed ? 'Permite' : 'Bloquea')}</td>
              <td>${formatDate(item.expires_at)}</td>
              <td><button class="button danger" data-click="deleteOverride" data-id="${item.id}">Eliminar</button></td>
            </tr>`).join('')}
        </tbody>
      </table>
    </div>`;
}

function rbacView() {
  return `
    <div class="topbar">
      <div><h1>RBAC</h1><p>Roles, modulos y features precalculados en JWT.</p></div>
      <button class="button" data-click="rolesTab" data-tab="${state.tab}">Actualizar vista</button>
    </div>
    <section class="panel">
      <div class="tabs">
        ${['roles', 'modules'].map((tab) => `<button class="${state.tab === tab ? 'active' : ''}" data-click="rolesTab" data-tab="${tab}">${tab === 'roles' ? 'Roles' : 'Modulos y features'}</button>`).join('')}
      </div>
      ${state.tab === 'roles' ? rolesView() : modulesView()}
    </section>`;
}

function rolesView() {
  return `
    <div class="table-wrap">
      <table>
        <thead><tr><th>Rol</th><th>Descripcion</th><th>Features</th><th>Tipo</th></tr></thead>
        <tbody>
          ${state.roles.map((role) => `
            <tr>
              <td><strong>${role.name}</strong><div class="muted">${role.display_name}</div></td>
              <td>${escapeHtml(role.description || '')}</td>
              <td>${badge('blue', `${role.features.length} permisos`)}</td>
              <td>${badge(role.is_system_role ? 'green' : 'gray', role.is_system_role ? 'Sistema' : 'Custom')}</td>
            </tr>`).join('')}
        </tbody>
      </table>
    </div>`;
}

function modulesView() {
  return `
    <div class="panel-body grid">
      ${state.modules.map((module) => `
        <article class="panel" style="box-shadow:none">
          <div class="panel-header">
            <div><h2 class="panel-title">${module.code}</h2><p class="panel-subtitle">${escapeHtml(module.name)} · ${module.features.length} features</p></div>
            ${badge('gray', `Orden ${module.display_order}`)}
          </div>
          <div class="panel-body feature-list">
            ${module.features.map((feature) => badge(feature.action_level === 'READ' ? 'blue' : feature.action_level === 'PUBLISH' ? 'green' : 'yellow', `${feature.code}:${feature.action_level}`)).join('')}
          </div>
        </article>`).join('')}
    </div>`;
}

function auditView() {
  return `
    <div class="topbar">
      <div><h1>Auditoría de login</h1><p>Intentos exitosos, fallidos y bloqueos.</p></div>
      <button class="button" data-click="refreshAudit">Actualizar</button>
    </div>
    <form class="toolbar" data-action="auditFilters">
      <div class="field"><label>Email</label><input name="email" type="email" placeholder="usuario@sena.edu.co"></div>
      <div class="field"><label>Resultado</label><select name="outcome"><option value="">Todos</option><option>SUCCESS</option><option>INVALID_PASSWORD</option><option>USER_NOT_FOUND</option><option>ACCOUNT_LOCKED</option><option>TOKEN_EXPIRED</option></select></div>
      <button class="button blue" type="submit">Filtrar</button>
    </form>
    <section class="panel">
      <div class="table-wrap">
        <table>
          <thead><tr><th>Fecha</th><th>Email</th><th>Resultado</th><th>IP</th><th>User agent</th></tr></thead>
          <tbody>
            ${state.audit.map((row) => `
              <tr>
                <td>${formatDate(row.attempted_at)}</td>
                <td>${escapeHtml(row.email_attempted)}</td>
                <td>${badge(row.outcome === 'SUCCESS' ? 'green' : 'red', row.outcome)}</td>
                <td>${escapeHtml(row.ip_address || '-')}</td>
                <td>${escapeHtml(row.user_agent || '-')}</td>
              </tr>`).join('')}
          </tbody>
        </table>
      </div>
      ${state.audit.length ? '' : '<div class="empty">Aun no hay eventos.</div>'}
      ${state.pagination?.next_cursor ? `<div class="panel-body"><button class="button" data-click="nextAudit" data-cursor="${state.pagination.next_cursor}">Cargar mas</button></div>` : ''}
    </section>`;
}

function createUserModal() {
  return `
    <div class="modal-backdrop">
      <section class="modal">
        <div class="panel-header"><h2 class="panel-title">Nuevo usuario</h2><button class="button icon" data-click="closeModal">x</button></div>
        <form data-action="createUser">
          <div class="form-grid">
            ${field('Correo', 'email', 'email', 'nuevo.usuario@sena.edu.co')}
            ${field('Nombre', 'first_name', 'text', 'Laura')}
            ${field('Apellido', 'last_name', 'text', 'Ramirez')}
            <div class="field"><label>Tipo de actor</label><select name="actor_type"><option>USER</option><option>INSTRUCTOR</option><option>LEARNER</option></select></div>
            <div class="field"><label>Rol inicial</label><select name="initial_role"><option value="">Sin rol</option>${state.roles.map((r) => `<option>${r.name}</option>`).join('')}</select></div>
            ${field('Centro de formacion', 'training_center_id', 'text', TRAINING_CENTER_ID)}
            ${field('Actor ID', 'actor_id', 'text', '')}
          </div>
          <div class="modal-actions"><button class="button" type="button" data-click="closeModal">Cancelar</button><button class="button primary" type="submit">Crear</button></div>
        </form>
      </section>
    </div>`;
}

function assignRoleModal() {
  return `
    <div class="modal-backdrop">
      <section class="modal">
        <div class="panel-header"><h2 class="panel-title">Asignar rol</h2><button class="button icon" data-click="closeModal">x</button></div>
        <form data-action="assignRole">
          <div class="form-grid">
            <div class="field full"><label>Rol</label><select name="role_name">${state.roles.map((r) => `<option>${r.name}</option>`).join('')}</select></div>
            ${field('Centro de formacion', 'training_center_id', 'text', TRAINING_CENTER_ID)}
            ${field('Expira', 'expires_at', 'datetime-local', '')}
          </div>
          <div class="modal-actions"><button class="button" type="button" data-click="closeModal">Cancelar</button><button class="button primary" type="submit">Asignar</button></div>
        </form>
      </section>
    </div>`;
}

function editUserModal() {
  const user = state.selectedUser;
  return `
    <div class="modal-backdrop">
      <section class="modal">
        <div class="panel-header"><h2 class="panel-title">Editar usuario</h2><button class="button icon" data-click="closeModal">x</button></div>
        <form data-action="updateUser">
          <div class="form-grid">
            ${field('Nombre', 'first_name', 'text', user.first_name)}
            ${field('Apellido', 'last_name', 'text', user.last_name)}
            <div class="field"><label>Tipo de actor</label><select name="actor_type">${['USER', 'INSTRUCTOR', 'LEARNER'].map((v) => `<option ${v === user.actor_type ? 'selected' : ''}>${v}</option>`).join('')}</select></div>
            ${field('Actor ID', 'actor_id', 'text', user.actor_id || '')}
            <div class="field"><label>Estado</label><select name="is_active"><option value="true" ${user.is_active ? 'selected' : ''}>Activo</option><option value="false" ${!user.is_active ? 'selected' : ''}>Inactivo</option></select></div>
          </div>
          <div class="modal-actions"><button class="button" type="button" data-click="closeModal">Cancelar</button><button class="button primary" type="submit">Guardar</button></div>
        </form>
      </section>
    </div>`;
}

function scopeOverrideModal() {
  const features = state.modules.flatMap((module) => module.features.map((feature) => feature.code));
  return `
    <div class="modal-backdrop">
      <section class="modal">
        <div class="panel-header"><h2 class="panel-title">Override de scope</h2><button class="button icon" data-click="closeModal">x</button></div>
        <form data-action="createOverride">
          <div class="form-grid">
            <div class="field"><label>Feature</label><select name="feature_code">${features.map((f) => `<option>${f}</option>`).join('')}</select></div>
            <div class="field"><label>Scope</label><select name="scope_type">${['GLOBAL', 'TRAINING_CENTER', 'AREA', 'OWN_FICHAS', 'OWN_SCHEDULE', 'OWN_PROFILE', 'OWN_FICHA_AS_LEARNER'].map((s) => `<option>${s}</option>`).join('')}</select></div>
            <div class="field"><label>Tipo</label><select name="is_allowed"><option value="true">Permitir</option><option value="false">Bloquear</option></select></div>
            ${field('Expira', 'expires_at', 'datetime-local', '')}
            <div class="field full"><label>Justificacion</label><input name="reason" required value="Cobertura temporal autorizada"></div>
          </div>
          <div class="modal-actions"><button class="button" type="button" data-click="closeModal">Cancelar</button><button class="button primary" type="submit">Crear</button></div>
        </form>
      </section>
    </div>`;
}

async function ensureDashboardData() {
  await loadRbacOnce();
  await loadUsersOnce();
}

async function loadUsersOnce() {
  if (!state.users.length) await loadUsers();
  if (!state.roles.length) await loadRbacOnce();
}

async function loadRbacOnce() {
  if (state.roles.length && state.modules.length) return;
  await run(async () => {
    const [roles, modules] = await Promise.all([api('/roles'), api('/modules')]);
    state.roles = roles;
    state.modules = modules;
  }, false);
}

async function loadAuditOnce() {
  if (!state.audit.length) await loadAudit();
}

async function loadUsers(filters = {}) {
  await run(async () => {
    const params = new URLSearchParams({ page: '1', page_size: '20', sort: 'created_at:desc' });
    for (const [key, value] of Object.entries(filters)) {
      if (value) params.set(key, value);
    }
    const response = await api(`/users?${params}`);
    state.users = response.data;
    state.pagination = response.pagination;
    if (!state.selectedUser && state.users[0]) {
      await selectUser(state.users[0].id, false);
    }
    render();
  }, false);
}

async function loadAudit(filters = {}) {
  await run(async () => {
    const params = new URLSearchParams({ limit: '50' });
    for (const [key, value] of Object.entries(filters)) {
      if (value) params.set(key, value);
    }
    const response = await api(`/reports/login-audit?${params}`);
    state.audit = response.data;
    state.pagination = response.pagination;
    render();
  }, false);
}

async function selectUser(id, rerender = true) {
  await run(async () => {
    const [user, sessions, overrides] = await Promise.all([
      api(`/users/${id}`),
      api(`/users/${id}/sessions`),
      api(`/users/${id}/scope-overrides`).catch((error) => {
        if (error.status === 403) return [];
        throw error;
      })
    ]);
    state.selectedUser = user;
    state.sessions = sessions;
    state.overrides = overrides;
    if (rerender) render();
  }, false);
}

async function api(path, options = {}) {
  const headers = { 'Content-Type': 'application/json' };
  if (!options.skipAuth && state.auth?.access_token) {
    headers.Authorization = `Bearer ${state.auth.access_token}`;
  }
  const response = await fetch(`${API_BASE}${path}`, {
    method: options.method || 'GET',
    headers,
    body: options.body ? JSON.stringify(options.body) : undefined
  });

  if (response.status === 401 && state.auth?.refresh_token && !options.skipAuth && !path.includes('/auth/refresh')) {
    const refreshed = await fetch(`${API_BASE}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: state.auth.refresh_token })
    });
    if (refreshed.ok) {
      state.auth = await refreshed.json();
      localStorage.setItem('iam_auth', JSON.stringify(state.auth));
      return api(path, options);
    }
  }

  if (!response.ok) {
    let payload = {};
    try {
      payload = await response.json();
    } catch (_error) {
      payload = {};
    }
    const error = new Error(payload.error?.message || 'Error de API');
    error.status = response.status;
    error.payload = payload;
    throw error;
  }
  if (options.empty || response.status === 204 || response.status === 202) return null;
  return response.json();
}

async function run(work, rerender = true) {
  state.error = '';
  state.message = '';
  state.loading = true;
  try {
    await work();
  } catch (error) {
    state.error = error.message || 'Error inesperado';
    if (error.status === 401) {
      localStorage.removeItem('iam_auth');
      state.auth = null;
      state.me = null;
      location.hash = '#/login';
    }
  } finally {
    state.loading = false;
    if (rerender) render();
  }
}

function openModal(html) {
  document.body.insertAdjacentHTML('beforeend', html);
}

function closeModal() {
  document.querySelector('.modal-backdrop')?.remove();
}

function navigate(path) {
  location.hash = `#${path}`;
}

function isPublicRoute(route) {
  const path = String(route || '').split('?')[0];
  return path === '/forgot-password' || path === '/reset-password';
}

function getRouteParam(name) {
  const route = String(state.route || '');
  const queryIndex = route.indexOf('?');
  if (queryIndex === -1) return '';
  return new URLSearchParams(route.slice(queryIndex + 1)).get(name) || '';
}

function topbar() {
  return `
    <div class="topbar">
      <div><h1>SENA IAM</h1><p>${escapeHtml(state.me?.full_name || '')} · ${(state.me?.roles || []).map((r) => r.name).join(', ')}</p></div>
      <button class="button danger" data-click="logout">Cerrar sesión</button>
    </div>`;
}

function brand() {
  return `
    <a class="brand" href="#/dashboard">
      <span class="brand-mark">SENA</span>
      <span class="brand-copy"><strong>Gestión de Horarios</strong><span>Identidad y Seguridad</span></span>
    </a>`;
}

function navButton(path, label, active) {
  return `<button class="${active ? 'active' : ''}" onclick="location.hash='#${path}'">${label}</button>`;
}

function metric(label, value) {
  return `<article class="metric"><span>${label}</span><strong>${value}</strong></article>`;
}

function badge(color, label) {
  return `<span class="badge ${color}">${escapeHtml(String(label))}</span>`;
}

function field(label, name, type, value) {
  return `<div class="field"><label>${label}</label><input name="${name}" type="${type}" value="${escapeHtml(value || '')}"></div>`;
}

function messageBlock() {
  const parts = [];
  if (state.error) parts.push(`<div class="alert error">${escapeHtml(state.error)}</div>`);
  if (state.message) parts.push(`<div class="alert success">${escapeHtml(state.message)}</div>`);
  return parts.join('');
}

function formatDate(value) {
  if (!value) return '-';
  return new Intl.DateTimeFormat('es-CO', {
    dateStyle: 'short',
    timeStyle: 'short'
  }).format(new Date(value));
}

function escapeHtml(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

render();
