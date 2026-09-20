/* ===========================================================
   RepairVision V9 统一登录页逻辑
   - 登录接口：POST /api/login（与三端共用）
   - 登录成功按 role 跳转：1=管理员 /admin/，2=工人 /m/worker，3=宿管 /m/dorm
   - 登录态同时写入 localStorage（与 PC 管理端共享）与 sessionStorage（移动端同标签页）
   =========================================================== */
(function () {
  'use strict'

  var API_LOGIN = '/api/login'
  var LOCAL_KEYS = ['rv_token', 'rv_user', 'rv_role', 'rv_admin_auth', 'rv_admin_token']
  var SESSION_KEYS = ['rv_token', 'rv_user']

  var $ = function (id) { return document.getElementById(id) }
  var el = {
    form: $('loginForm'),
    signed: $('signed'),
    signedUser: $('signedUser'),
    continueBtn: $('continueBtn'),
    switchBtn: $('switchBtn'),
    username: $('username'),
    password: $('password'),
    togglePwd: $('togglePwd'),
    submitBtn: $('submitBtn'),
    alert: $('alert'),
    chips: $('demoChips')
  }

  /* ---------- 角色 ---------- */
  function targetFor(role) {
    if (role === 1) return '/admin/'
    if (role === 2) return '/m/worker'
    if (role === 3) return '/m/dorm'
    return ''
  }

  function roleText(role) {
    if (role === 1) return '超级管理员'
    if (role === 2) return '维修工人'
    if (role === 3) return '宿管'
    return '未知角色'
  }

  /* ---------- 登录态读写 ---------- */
  function readJSON(storage, key) {
    try {
      var raw = storage.getItem(key)
      return raw ? JSON.parse(raw) : null
    } catch (err) {
      return null
    }
  }

  function currentSession() {
    var token = sessionStorage.getItem('rv_token') || localStorage.getItem('rv_token')
    var user = readJSON(sessionStorage, 'rv_user') || readJSON(localStorage, 'rv_user')
    if (!token || !user) return null
    return { token: token, user: user }
  }

  function persist(token, user) {
    var userJson = JSON.stringify(user)
    localStorage.setItem('rv_token', token)
    localStorage.setItem('rv_user', userJson)
    localStorage.setItem('rv_role', String(user.role))
    localStorage.setItem('rv_admin_auth', JSON.stringify({ token: token, user: user }))
    localStorage.setItem('rv_admin_token', token)
    sessionStorage.setItem('rv_token', token)
    sessionStorage.setItem('rv_user', userJson)
  }

  function clearAll() {
    LOCAL_KEYS.forEach(function (key) { localStorage.removeItem(key) })
    SESSION_KEYS.forEach(function (key) { sessionStorage.removeItem(key) })
  }

  /* ---------- UI 辅助 ---------- */
  function showAlert(message) {
    el.alert.textContent = message
    el.alert.hidden = !message
  }

  function setLoading(loading) {
    el.submitBtn.disabled = loading
    el.submitBtn.classList.toggle('loading', loading)
    el.submitBtn.querySelector('.btn-text').textContent = loading ? '登录中…' : '登 录'
  }

  function showSigned(session) {
    el.signedUser.textContent = '当前账号：' + (session.user.name || session.user.username) +
      '（' + roleText(session.user.role) + '）'
    el.signed.hidden = false
    el.form.hidden = true
  }

  function showForm() {
    el.signed.hidden = true
    el.form.hidden = false
    el.form.reset()
    showAlert('')
    if (el.username) el.username.focus()
  }

  /* ---------- 登录 ---------- */
  function login(username, password) {
    return fetch(API_LOGIN, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username: username, password: password })
    }).then(function (resp) {
      return resp.json().catch(function () { return {} }).then(function (body) {
        if (!resp.ok || (body && typeof body.code === 'number' && body.code !== 0)) {
          throw new Error((body && body.msg) || ('登录失败（HTTP ' + resp.status + '）'))
        }
        var data = (body && body.data) || {}
        if (!data.token || !data.user) throw new Error('登录响应缺少 token 或用户信息')
        return data
      })
    })
  }

  function onSubmit(event) {
    event.preventDefault()
    var username = (el.username.value || '').trim()
    var password = el.password.value || ''
    if (!username || !password) {
      showAlert('请输入账号和密码')
      return
    }
    showAlert('')
    setLoading(true)
    login(username, password).then(function (data) {
      var target = targetFor(data.user.role)
      if (!target) {
        clearAll()
        throw new Error('该账号角色暂不支持网页端登录，请联系管理员')
      }
      persist(data.token, data.user)
      showAlert('')
      el.submitBtn.querySelector('.btn-text').textContent = '登录成功，正在进入…'
      window.setTimeout(function () { window.location.href = target }, 180)
    }).catch(function (err) {
      setLoading(false)
      showAlert(err && err.message ? err.message : '网络异常，请稍后重试')
    })
  }

  /* ---------- 事件绑定 ---------- */
  el.form.addEventListener('submit', onSubmit)

  el.togglePwd.addEventListener('click', function () {
    var showing = el.password.type === 'text'
    el.password.type = showing ? 'password' : 'text'
    el.togglePwd.textContent = showing ? '显示' : '隐藏'
    el.password.focus()
  })

  el.chips.addEventListener('click', function (event) {
    var chip = event.target.closest('.chip')
    if (!chip) return
    el.username.value = chip.getAttribute('data-username') || ''
    el.password.value = 'admin123'
    showAlert('')
    el.submitBtn.focus()
  })

  el.continueBtn.addEventListener('click', function () {
    var session = currentSession()
    var target = session ? targetFor(session.user.role) : ''
    if (target) window.location.href = target
    else showForm()
  })

  el.switchBtn.addEventListener('click', function () {
    clearAll()
    showForm()
  })

  /* ---------- 初始化 ---------- */
  var session = currentSession()
  if (session && targetFor(session.user.role)) {
    showSigned(session)
  } else {
    if (session) clearAll()
    showForm()
  }
})()