export default {
    beforeEach(to, from, next) {
        if (to.matched.some(record => record.meta.public)) {
            next();
        } else {
            axios
                .get('/api/v1/info')
                .then(r => {
                    if (r.data.data.type === 'user') {
                        next();
                    } else {
                        next({name: 'signin', params: {nextUrl: to.fullPath}});
                    }
                })
                .catch(e => {
                    console.log(e);
                });
        }
    },
    render(h) {
        return h('div');
    }
}
