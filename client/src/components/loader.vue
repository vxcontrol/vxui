<template>
    <div>
        <div class="outer" style="height:100px; width: 100%;">
      <span class="inner">
        <atom-spinner :animation-duration="1000" :size="60" :color="'#195f9b'"/>
      </span>
        </div>
    </div>
</template>
<script>
import {AtomSpinner} from "epic-spinners";

export default {
    data() {
        return {};
    },
    methods: {},
    created() {
    },
    components: {
        AtomSpinner
    }
};
</script>
<style scoped>
.outer:before {
    content: "";
    display: inline-block;
    height: 100%;
    vertical-align: middle;
}

.inner {
    display: inline-block;
    vertical-align: middle;
}

.outer {
    text-align: center;
}
</style>
