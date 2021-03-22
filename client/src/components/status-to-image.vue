<template>
    <img class="uk-border-circle" width="10" height="10" :src="detect">
</template>
<script>
export default {
    name: "status-to-image",
    props: ["status"],
    data() {
        return {
            imgPath: ""
        };
    },
    computed: {
        detect() {
            if (/disconnected/i.test(this.status)) {
                return "/images/status/red.png";
            } else if (/connected/i.test(this.status)) {
                return "/images/status/green.png";
            }
            return "/images/status/unknown.png";
        }
    }
};
</script>
