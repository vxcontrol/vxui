export default {
    getFilterObjs(fields, query) {
        const isContains = (val) => {
            if (typeof (val) !== "string") {
                val = JSON.stringify(val)
            }
            return val.toLowerCase().includes(query.toLowerCase());
        };
        return obj => {
            return fields.filter(field => {
                const opts = field.split(".");
                if (opts.length == 4) {
                    return isContains(obj[opts[0]][opts[1]][opts[2]][opts[3]]);
                } else if (opts.length == 3) {
                    return isContains(obj[opts[0]][opts[1]][opts[2]]);
                } else if (opts.length == 2) {
                    return isContains(obj[opts[0]][opts[1]]);
                } else {
                    return isContains(obj[field]);
                }
            }).length !== 0;
        }
    }
}
