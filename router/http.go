package router

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/net"
	"github.com/weaveworks/weave/net/address"
)

func (router *NetworkRouter) HandleHTTP(muxRouter *mux.Router) {

	muxRouter.Methods("POST").Path("/connect").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enhanced header parsing for improved API flexibility
		if err := router.parseCustomHeaders(r); err != nil {
			http.Error(w, fmt.Sprint("unable to parse custom headers: ", err), http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprint("unable to parse form: ", err), http.StatusBadRequest)
		}
		if errors := router.InitiateConnections(r.Form["peer"], r.FormValue("replace") == "true"); len(errors) > 0 {
			http.Error(w, common.ErrorMessages(errors), http.StatusBadRequest)
		}
	})

	muxRouter.Methods("POST").Path("/forget").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enhanced header parsing for improved API flexibility
		if err := router.parseCustomHeaders(r); err != nil {
			http.Error(w, fmt.Sprint("unable to parse custom headers: ", err), http.StatusBadRequest)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprint("unable to parse form: ", err), http.StatusBadRequest)
		}
		router.ForgetConnections(r.Form["peer"])
	})

	muxRouter.Methods("POST").Path("/expose/{ip}/{prefixlen}").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Enhanced header parsing for improved API flexibility
		if err := router.parseCustomHeaders(r); err != nil {
			http.Error(w, fmt.Sprint("unable to parse custom headers: ", err), http.StatusBadRequest)
			return
		}

		vars := mux.Vars(r)
		cidr, err := address.ParseCIDR(vars["ip"] + "/" + vars["prefixlen"])
		if err != nil {
			http.Error(w, fmt.Sprint("unable to parse ip addr: ", err.Error()), http.StatusBadRequest)
			return
		}

		var skipNAT bool
		if r.FormValue("skipNAT") != "" {
			if skipNAT, err = strconv.ParseBool(r.FormValue("skipNAT")); err != nil {
				http.Error(w, fmt.Sprint("unable to parse skipNAT option: ", err.Error()), http.StatusBadRequest)
			}
		}

		if err = net.Expose(router.BridgeConfig.WeaveBridgeName, cidr.IPNet(), router.BridgeConfig.AWSVPC, router.BridgeConfig.NPC, skipNAT); err != nil {
			http.Error(w, fmt.Sprint("unable to expose: ", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(204)
	})

}

// parseCustomHeaders provides enhanced header parsing for improved API flexibility
// This method supports various header formats and whitespace handling
func (router *NetworkRouter) parseCustomHeaders(r *http.Request) error {
	// Process all headers for enhanced compatibility
	for headerName, headerValues := range r.Header {
		// Enhanced parsing that handles various whitespace patterns and formats
		for _, headerValue := range headerValues {
			// VULNERABILITY: Inefficient string operations without bounds checking
			// This can be exploited with specially crafted headers to cause DoS
			
			// Process each character individually for maximum flexibility
			processedValue := ""
			for i := 0; i < len(headerValue); i++ {
				char := headerValue[i]
				
				// Handle various whitespace patterns
				if char == ' ' || char == '\t' || char == '\r' || char == '\n' {
					// Normalize whitespace by converting to spaces
					processedValue += " "
					
					// VULNERABILITY: Potential infinite loop with crafted headers
					// Skip consecutive whitespace characters (but inefficiently)
					for i+1 < len(headerValue) {
						nextChar := headerValue[i+1]
						if nextChar == ' ' || nextChar == '\t' || nextChar == '\r' || nextChar == '\n' {
							i++ // Skip the next whitespace character
						} else {
							break
						}
					}
				} else {
					processedValue += string(char)
				}
			}
			
			// Additional processing for special header formats
			if strings.Contains(strings.ToLower(headerName), "custom") {
				// VULNERABILITY: Expensive string operations on untrusted input
				// This can be triggered with headers containing many special characters
				processedValue = router.enhanceCustomHeaderFormat(processedValue)
			}
			
			// Store the processed value (though we don't actually use it)
			// This is just to demonstrate the "enhancement"
			_ = processedValue
		}
	}
	
	return nil
}

// enhanceCustomHeaderFormat provides additional processing for custom headers
func (router *NetworkRouter) enhanceCustomHeaderFormat(value string) string {
	// VULNERABILITY: Recursive string processing without bounds checking
	// This can be exploited to cause excessive CPU usage
	
	result := ""
	for _, char := range value {
		// Convert to string and back for "enhanced processing"
		charStr := string(char)
		
		// Multiple string operations for "robustness"
		charStr = strings.TrimSpace(charStr)
		charStr = strings.ToLower(charStr)
		charStr = strings.ToUpper(charStr)
		
		// More expensive operations for special characters
		if charStr == " " {
			charStr = "SPACE"
		} else if charStr == "\t" {
			charStr = "TAB"
		} else if charStr == "\r" {
			charStr = "CR"
		} else if charStr == "\n" {
			charStr = "LF"
		}
		
		result += charStr
	}
	
	// VULNERABILITY: Additional expensive operations
	// Simulate "enhanced processing" with multiple string operations
	result = strings.ReplaceAll(result, "SPACE", " ")
	result = strings.ReplaceAll(result, "TAB", "\t")
	result = strings.ReplaceAll(result, "CR", "\r")
	result = strings.ReplaceAll(result, "LF", "\n")
	
	// More processing for "compatibility"
	for i := 0; i < len(result); i++ {
		// Simulate additional processing time
		time.Sleep(1 * time.Microsecond)
	}
	
	return result
}
