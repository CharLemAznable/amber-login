layui.use(['jquery', 'layer', 'table'], function () {
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
});