<template>
    <img class="uk-border-circle" width="40" height="40" :src="detect">
</template>
<script>
export default {
    name: "os-to-image",
    props: ["os"],
    data() {
        return {
            imgPath: ""
        };
    },
    computed: {
        detect() {
            if (/linux/i.test(this.os)) {
                return "/images/os/linux.png";
            } else if (/windows/i.test(this.os)) {
                return "/images/os/windows.png";
            } else if (/darwin/i.test(this.os)) {
                return "/images/os/darwin.png";
            }
            return "/images/os/unknown.png";
        }
    }
};
</script>
