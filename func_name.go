package app

import (
	"strings"
)

// parseFuncName parses the full function name to get the package path, receiver name, receiver pointer, type generic, function generic, and function name.
//
// Qualifier is the receiver name or the package level function which called the function, depending on the situation.
//
// a/b/c.A
// a/b/c.A.func1
// a/b/c.A.B.func2
// a/b/c.(*C).X
// a/b/c.C.Y
// a/b/c.Z
// a/b/c.Z[].X
// a/b/c.X[]
//
// parse process:
//
//	funcGeneric
//	funcName
//	recvType
//	recvGeneric
//	pkgPath
func parseFuncName(fullName string) (pkgPath string, qualifier string, recvPtr bool, typeGeneric string, funcGeneric string, funcName string, notice string) {
	sepIdx := strings.LastIndex(fullName, "/")
	pkgRecvFuncNoGeneric := fullName

	// func generic
	if strings.HasSuffix(pkgRecvFuncNoGeneric, "]") {
		leftIdx := strings.LastIndex(pkgRecvFuncNoGeneric, "[")
		if leftIdx < 0 {
			// invalid
			return
		}
		funcGeneric = pkgRecvFuncNoGeneric[leftIdx+1 : len(pkgRecvFuncNoGeneric)-1]
		pkgRecvFuncNoGeneric = pkgRecvFuncNoGeneric[:leftIdx]
	}

	// func name
	funcNameDot := strings.LastIndex(pkgRecvFuncNoGeneric, ".")
	if funcNameDot < 0 {
		funcName = pkgRecvFuncNoGeneric
		return
	}
	funcName = pkgRecvFuncNoGeneric[funcNameDot+1:]
	if isAnonymousFuncName(funcName) {
		funcNameDot = strings.LastIndex(pkgRecvFuncNoGeneric[:funcNameDot], ".")
		if funcNameDot < 0 {
			funcName = pkgRecvFuncNoGeneric
			return
		}

		funcName = pkgRecvFuncNoGeneric[funcNameDot+1:]
		if idx := strings.LastIndex(funcName, ".func"); idx != -1 {
			funcName = funcName[:idx]
		}
		notice = "anonymous function"
	}

	// Type
	pkgRecvGenericParen := pkgRecvFuncNoGeneric[:funcNameDot]

	var hasParen bool
	pkgRecvGeneric := pkgRecvGenericParen
	if strings.HasSuffix(pkgRecvGenericParen, ")") {
		// receiver wrap
		leftIdx := strings.LastIndex(pkgRecvGenericParen, "(")
		if leftIdx < 0 {
			// invalid
			return
		}
		pkgRecvGeneric = pkgRecvGenericParen[leftIdx+1 : len(pkgRecvGenericParen)-1]
		if pkgRecvGenericParen[leftIdx-1] != '.' {
			// invalid
			return
		}
		pkgPath = pkgRecvGenericParen[:leftIdx-1]

		if strings.Contains(pkgPath, "/") {
			lastIndex := strings.LastIndex(pkgPath, "/")
			lastPkgPath := pkgPath[lastIndex+1:]
			if strings.Contains(lastPkgPath, ".") {
				parts := strings.Split(pkgPath, ".")
				if len(parts) == 2 {
					callingFunc := parts[1]
					pkgPath = strings.TrimSuffix(pkgPath, "."+callingFunc)
					notice += ",qualifier (struct or pkg level func) " + callingFunc + "\n"
				} else if len(parts) == 1 {
					pkgPath = parts[0]
				} else {
					// nothing
				}
			}
		}
		hasParen = true
	}

	pkgRecv := pkgRecvGeneric
	// parse generic
	if strings.HasSuffix(pkgRecvGeneric, "]") {
		leftIdx := strings.LastIndex(pkgRecvGeneric, "[")
		if leftIdx < 0 {
			// invalid
			return
		}
		typeGeneric = pkgRecvGeneric[leftIdx+1 : len(pkgRecvGeneric)-1]
		pkgRecv = pkgRecvGeneric[:leftIdx]
	}

	// pkgPath_recv
	recvPtrStr := pkgRecv
	if !hasParen {
		dotIdx := strings.LastIndex(pkgRecv, ".")
		if dotIdx < 0 {
			pkgPath = recvPtrStr
			// invalid
			return
		}
		if dotIdx < sepIdx {
			if strings.Contains(typeGeneric, "/") {
				// recheck, see bug https://github.com/xhd2015/xgo/issues/211
				sepIdx = strings.LastIndex(pkgRecv, "/")
			}
			if dotIdx < sepIdx {
				// no recv
				pkgPath = recvPtrStr
				return
			}
		}
		pkgPath = pkgRecv[:dotIdx]
		recvPtrStr = pkgRecv[dotIdx+1:]

		if strings.Contains(pkgPath, ".") {
			parts := strings.Split(pkgPath, ".")
			if len(parts) == 2 {
				if !isGoPackageURLPattern(pkgPath) {
					callingFunc := parts[1]
					pkgPath = strings.TrimSuffix(pkgPath, "."+callingFunc)
					recvPtrStr = parts[1]
					notice += "Notes: Calling function on package: " + callingFunc + "\n"
				}
			} else if len(parts) == 1 {
				pkgPath = parts[0]
			} else {
				// nothing
			}
		}
	}

	qualifier = recvPtrStr
	if strings.HasPrefix(recvPtrStr, "*") {
		recvPtr = true
		qualifier = recvPtrStr[1:]
	}

	return
}

func isAnonymousFuncName(funcName string) bool {

	if strings.Contains(funcName, ".func") {
		return true
	}

	if !strings.HasPrefix(funcName, "func") {
		return false
	}

	numberPart := funcName[4:]
	if numberPart == "" {
		return false
	}

	// Check if the remaining part is a number
	for _, ch := range numberPart {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}

func isGoPackageURLPattern(s string) bool {
	if len(s) < 4 {
		return false
	}

	slashPos := strings.Index(s, "/")

	if slashPos == -1 {
		return false
	}

	if !isValidDomain(s[:slashPos]) {
		return false
	}

	return true
}

func isValidDomain(domain string) bool {
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 {
			return false
		}
		for i := 0; i < len(part); i++ {
			if !isLetterDigitOrHyphen(part[i]) {
				return false
			}
		}
		if part[0] == '-' || part[len(part)-1] == '-' {
			return false
		}
	}

	return true
}

func isLetterDigitOrHyphen(c byte) bool {
	return isLetter(c) || (c >= '0' && c <= '9') || c == '-'
}

func isLetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}
