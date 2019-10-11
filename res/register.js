layui.use(['jquery', 'layer', 'form'], function () {
    let $ = layui.$;
    let layer = layui.layer;
    let form = layui.form;

    form.verify({
        pass: [
            /^(?![0-9]+$)(?![a-zA-Z]+$)[0-9A-Za-z]{6,20}$/,
            '密码必须为6-20位, 必须包含字母和数字'
        ]
    });

    layer.open({
        type: 1,
        title: '用户注册',
        content: $('.register-content'),
        area: '300px',
        btn: [],
        btnAlign: 'c',
        closeBtn: false,
        shade: 0.6,
        shadeClose: false,
        id: 'register-content',
        resize: false,
        move: false,
    });

    $(document).on('keydown', function (event) {
        // noinspection JSDeprecatedSymbols
        if (event.keyCode === 13) {
            if ($(document).find('.layui-layer').length === 1) {
                $('.register-btn').click();
            } else {
                layer.close(layer.index);
            }
        }
    });

    form.on('submit(register)', function (data) {
        let load = layer.load(1, {shade: 0.6});
        $.ajax({
            type: 'POST',
            url: '${contextPath}/do-register',
            contentType: 'application/json',
            data: JSON.stringify(data.field),
            dataType: 'json',
            success: function (res) {
                if ('OK' !== res.msg) {
                    if ('验证码不存在或已过期' === res.msg) {
                        layer.alert(res.msg, function () {
                            location.replace(location.href);
                        });
                        return;
                    }
                    layer.alert(res.msg);
                    return;
                }
                layer.open({
                    type: 1,
                    content: '<div style="margin:10px">注册成功, 请等待管理员审核...</div>',
                    btn: ['确认'],
                    btnAlign: 'c',
                    closeBtn: false,
                    shadeClose: false,
                    id: 'register-success',
                    resize: false,
                    move: false,
                    yes: function () {
                        location.replace(location.href);
                    },
                });
            },
            error: function () {
                layer.alert('注册异常');
            },
            complete: function () {
                layer.close(load);
            },
        });
        return false;
    });
});
