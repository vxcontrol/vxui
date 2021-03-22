import Vue from 'vue';
import VueI18n from 'vue-i18n';
import { upperFirst } from 'lodash-es';

import element_ui_locale_en from 'element-ui/lib/locale/lang/en';
import element_ui_locale_ru from 'element-ui/lib/locale/lang/ru-RU';

import enFeatures from '@/locales/en-US/features.json';
import ruFeatures from '@/locales/ru-RU/features.json';
import enCommon from '@/locales/en-US/common.json';
import ruCommon from '@/locales/ru-RU/common.json';
import enApp from '@/locales/en-US/app.json';
import ruApp from '@/locales/ru-RU/app.json';


function VuI18nExt() {
}

VuI18nExt.install = function () {
    // локализация композитных ключей (статическая часть + значение, к примеру, с backend'а)
    Vue.prototype.$tck = function (prefix, value) {
        return this.$root.$t(prefix + toLocalizationKeyPart(value));
    };
}

Vue.use(VueI18n);
Vue.use(VuI18nExt);

const availableLocales = [ 'ru', 'en' ];
const locale = getLocale();
const messages = { en: { ...enFeatures, ...enApp, ...enCommon }, ru: { ...ruFeatures, ...ruApp, ...ruCommon } };
const elementUiMessages = { en: element_ui_locale_en, ru: element_ui_locale_ru };

const i18n = new VueI18n({ locale, messages });

function getLocale() {
    return localStorage.getItem('locale') || 'en';
}

function setLocale(locale) {
    i18n.locale = locale;
    localStorage.setItem('locale', locale);
}

function toLocalizationKeyPart(inputString) {
    const value = inputString
        // заменяем букву после не-буквы на заглавную
        .replace(/[-_.](\w)/ig, (match, p1) => upperFirst(p1))
        // удаляем не-буквы и не-цифры
        .replace(/[^a-z0-9\d]/ig, '');

    return upperFirst(value);
}


export {
    i18n,
    availableLocales,
    elementUiMessages,
    getLocale,
    setLocale,
    toLocalizationKeyPart
}
