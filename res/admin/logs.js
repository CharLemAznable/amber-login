layui.use(['jquery', 'layer', 'form', 'table'], function () {
    let $ = layui.$;
    let layer = layui.layer;
    let form = layui.form;
    let table = layui.table;

    let tpl = function (d) {
        return d['app-name'] + '[' + d['app-id'] + ']';
    };

    let col = [
        {field: 'username', title: '用户名', width: '20%'},
        {field: 'app-name', title: '登录应用', width: '20%', 'templet': tpl},
        {field: 'login-time', title: '登录时间', width: '20%'},
        {field: 'redirect-url', title: '跳转地址', width: '40%'},
    ];

    let loadLogs = function () {
        table.render({
            elem: '#logs',
            cols: [col],
            page: true,
            limit: 10,
            limits: [10, 20, 50],
            url: '${contextPath}/admin/query-logs',
        });
    };
    loadLogs();

    form.on('submit(clean)', function (data) {
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/admin/clean-logs',
            contentType: 'application/json',
            data: JSON.stringify(data.field),
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    layer.alert(res.msg, function () {
                        location.replace(location.href);
                    });
                    return;
                }
                location.replace(location.href);
            },
            error: function () {
                layer.alert('服务异常');
            },
            complete: function () {
                layer.close(load);
            },
        });
        return false;
    });
});