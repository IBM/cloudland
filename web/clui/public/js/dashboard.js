import i18n from '@/assets/i18n'
if (i18n.locale == 'zh') {
    $.getScript('/js/dashboard-zh.js');
} else {
    $.getScript('/js/dashboard-en.js');
}
