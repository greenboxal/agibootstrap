/*const ParsePathEx = (s: string, escaped: boolean): Path => {
    const rootSepIndex = s.indexOf('//');
    const isRelative = rootSepIndex === -1;
    const root = isRelative ? '' : s.slice(0, rootSepIndex);
    const path = isRelative ? s : s.slice(rootSepIndex + 2);
    const components = ParsePathElements(path, escaped);

    if (!components) {
        throw new Error('Error parsing path elements');
    }

    return PathFromElements(root, isRelative, ...components);
};



const PathFromElements = (root: string, isRelative: boolean, ...components: PathElement[]): Path => {
    return new Path(isRelative, root, [...components]);
};*/

type PathElement = {
    kind: string;
    name: string;
    index: number;
}

export const parsePathElements = (s: string, escaped: boolean): PathElement[] | null => {
    s = s.trim();
    s = s.replace(/\/+$/, '');

    if (s === '') {
        return null;
    }

    const sp = s.split('/');
    const result: PathElement[] = [];

    for (let i = 0; i < sp.length; i++) {
        const part = sp[i].trim();

        if (part === '') {
            if (i === 0) {
                continue;
            } else {
                throw new Error('Empty path component');
            }
        }

        if (escaped) {
            // Note: decodeURIComponent should handle most cases, but may not be a perfect match
            try {
                decodeURIComponent(part);
            } catch (err) {
                throw new Error('Error decoding URI component');
            }
        }

        const element = parsePathElement(part);
        result.push(element);
    }

    return result;
};

export const parsePathElement = (str: string): PathElement => {
    let state = '#';
    let acc = '';
    const element: PathElement = {
        kind: '',
        name: '',
        index: 0,
    }

    for (const ch of str) {
        switch (ch) {
            case '@':
            case '#':
            case ':':
                if (acc !== '') {
                    switch (state) {
                        case '@':
                            element.index = parseInt(acc);
                            break;
                        case '#':
                            element.name = acc;
                            break;
                        case ':':
                            element.kind = acc;
                            break;
                    }
                }
                acc = '';
                state = ch;
                break;
            default:
                acc += ch;
                break;
        }
    }

    if (acc !== '') {
        switch (state) {
            case '@':
                element.index = parseInt(acc);
                break;
            case '#':
                element.name = acc;
                break;
            case ':':
                element.kind = acc;
                break;
        }
    }

    return element;
};
