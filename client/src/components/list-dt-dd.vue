<template>
    <dl class="uk-description-list uk-description-list-divider">
        <template v-for="(e, key) in content">
            <dt :key="'t'+key">{{ e.title }}</dt>
            <dd :key="'d'+key">{{ e.description }}</dd>
        </template>
    </dl>
</template>
<script>
export default {
    props: {
        content: {}
    }
};
</script>
