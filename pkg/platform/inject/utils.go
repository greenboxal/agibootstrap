package inject

import `reflect`

var errorType = reflect.TypeOf((*error)(nil)).Elem()
var resolutionContextType = reflect.TypeOf((*ResolutionContext)(nil)).Elem()
