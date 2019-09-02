layui.use(['jquery', 'layer', 'table', 'form', 'transfer'], function () {
    let $ = layui.$;
    let layer = layui.layer;
    let table = layui.table;
    let form = layui.form;
    let transfer = layui.transfer;

    let col = [
        {field: 'username', title: '用户名', width: '20%'},
        {field: 'create-time', title: '注册时间', width: '20%'},
        {field: 'update-time', title: '更新时间', width: '20%'},
        {toolbar: '.user-toolbar', title: '操作'},
    ];

    let load = layer.load(1, {shade: 0.6});
    $.ajax({
        url: '${contextPath}/admin/query-users',
        dataType: 'json',
        success: function (res) {
            if ('OK' !== res.msg) {
                layer.alert(res.msg);
                return;
            }

            table.render({
                elem: '#users',
                cols: [col],
                data: res['users'],
                page: true,
                limit: 10,
                limits: [10, 20, 50],
                even: true,
            });
        },
        error: function () {
            layer.alert('查询失败');
        },
        complete: function () {
            layer.close(load);
        },
    });

    let privilegeSetting = function (data) {
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            url: '${contextPath}/admin/query-app-transfers',
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    layer.alert(res.msg);
                    return;
                }

                transfer.render({
                    elem: '.privilege-content .privilege-transfer',
                    title: ['无权限的应用', '有权限的应用'],
                    data: res['apps'],
                    parseData: function (item) {
                        return {
                            'value': item.id,
                            'title': item.name,
                            'disabled': item.disabled,
                            'checked': item.checked,
                        };
                    },
                    value: data['app-ids'],
                    id: 'privilege-transfer',
                });
                form.val('privilege-form', {
                    'username': data['username'],
                });
                layer.open({
                    type: 1,
                    title: '权限设置',
                    content: $('.privilege-content'),
                    area: ['500px', '500px'],
                    btn: [],
                    btnAlign: 'c',
                    closeBtn: 1,
                    shade: 0.6,
                    shadeClose: false,
                    id: 'privilege-form',
                    resize: false,
                    move: false,
                });
            },
            error: function () {
                layer.alert('查询失败');
            },
            complete: function () {
                layer.close(load);
            },
        });
    };

    form.on('submit(privilege)', function (data) {
        let username = data.field['username'];
        let transferData = transfer.getData('privilege-transfer');
        let appIds = transferData.map(x => x.value);

        let postData = {username: username, "app-ids": appIds};
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/admin/set-user-privileges',
            contentType: 'application/json',
            data: JSON.stringify(postData),
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    layer.alert(res.msg);
                    return;
                }
                layer.alert('操作成功', function () {
                    location.replace(location.href);
                });
            },
            error: function () {
                layer.alert('操作失败');
            },
            complete: function () {
                layer.close(load);
            },
        });
        return false;
    });

    let resetPassword = function (data) {
        form.val('reset-form', {
            'username': data['username'],
        });
        layer.open({
            type: 1,
            title: '重置密码',
            content: $('.reset-content'),
            area: '300px',
            btn: [],
            btnAlign: 'c',
            closeBtn: 1,
            shade: 0.6,
            shadeClose: false,
            id: 'reset-form',
            resize: false,
            move: false,
        });
    };

    form.on('submit(reset)', function (data) {
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/admin/reset-user-password',
            contentType: 'application/json',
            data: JSON.stringify(data.field),
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    layer.alert(res.msg);
                    return;
                }
                layer.alert('操作成功', function () {
                    location.replace(location.href);
                });
            },
            error: function () {
                layer.alert('操作失败');
            },
            complete: function () {
                layer.close(load);
            },
        });
        return false;
    });

    let switchToggle = function (data) {
        layer.confirm('确认' + (data['available'] ? '停用' : '启用') + '?', function () {
            let load = layer.load(1, {shade: 0.6});
            $.ajax({
                type: 'POST',
                url: '${contextPath}/admin/switch-toggle-user',
                contentType: 'application/json',
                data: JSON.stringify({username: data['username']}),
                dataType: 'json',
                success: function (res) {
                    if ('OK' !== res.msg) {
                        layer.alert(res.msg);
                        return;
                    }
                    layer.alert('操作成功', function () {
                        location.replace(location.href);
                    });
                },
                error: function () {
                    layer.alert('操作失败');
                },
                complete: function () {
                    layer.close(load);
                },
            });
        });
    };

    let deleteUser = function (data) {
        layer.confirm('确认删除?', function () {
            let load = layer.load(1, {shade: 0.6});
            $.ajax({
                type: 'POST',
                url: '${contextPath}/admin/delete-user',
                contentType: 'application/json',
                data: JSON.stringify({username: data['username']}),
                dataType: 'json',
                success: function (res) {
                    if ('OK' !== res.msg) {
                        layer.alert(res.msg);
                        return;
                    }
                    layer.alert('操作成功', function () {
                        location.replace(location.href);
                    });
                },
                error: function () {
                    layer.alert('操作失败');
                },
                complete: function () {
                    layer.close(load);
                },
            });
        });
    };

    table.on('tool(users)', function (obj) {
        let data = obj.data;
        switch (obj.event) {
            case 'privilege':
                privilegeSetting(data);
                break;
            case 'reset':
                resetPassword(data);
                break;
            case 'switch':
                switchToggle(data);
                break;
            case 'delete':
                deleteUser(data);
                break;
        }
    });

    let address = 'ws://' + location.host + '${contextPath}/admin/users/websocket';
    let socket = new WebSocket(address);
    socket.onopen = function () {
        socket.send(decodeURIComponent(document.cookie));
    };
    socket.onmessage = function (e) {
        let res = JSON.parse(e.data);
        table.render({
            elem: '#users',
            cols: [col],
            data: res['users'],
            page: true,
            limit: 10,
            limits: [10, 20, 50],
            even: true,
        });
    }
});