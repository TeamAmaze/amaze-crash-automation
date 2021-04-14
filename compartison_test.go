package main

import (
    "fmt"
    "testing"
    "strings"
    "hash/fnv"
)

const test1A = `java.lang.ArrayIndexOutOfBoundsException: length=366; index=-1
at java.util.ArrayList.get(ArrayList.java:439)
at
com.amaze.filemanager.adapters.RecyclerAdapter.toggleChecked(RecyclerAdapter.java:179)
at
com.amaze.filemanager.ui.fragments.MainFragment.onListItemClicked(MainFragment.java:799)
at
com.amaze.filemanager.adapters.RecyclerAdapter.lambda$onBindViewHolder$0(RecyclerAdapter.java:552)
at
com.amaze.filemanager.adapters.RecyclerAdapter.lambda$onBindViewHolder$0$RecyclerAdapter(Unknown
Source:0)
at
com.amaze.filemanager.adapters.-$$Lambda$RecyclerAdapter$FDaxlO2LtK_Q0Z-YEozjkaXX_T8.onClick(Unknown
Source:11)
at android.view.View.performClick(View.java:6748)
at android.view.View$PerformClick.run(View.java:25458)
at android.os.Handler.handleCallback(Handler.java:790)
at android.os.Handler.dispatchMessage(Handler.java:99)
at android.os.Looper.loop(Looper.java:164)
at android.app.ActivityThread.main(ActivityThread.java:6549)
at java.lang.reflect.Method.invoke(Native Method)
at
com.android.internal.os.RuntimeInit$MethodAndArgsCaller.run(RuntimeInit.java:438)
at com.android.internal.os.ZygoteInit.main(ZygoteInit.java:888)`

const test1B = `java.lang.ArrayIndexOutOfBoundsException: length=22; index=-1
at java.util.ArrayList.get(ArrayList.java:439)
at com.amaze.filemanager.adapters.RecyclerAdapter.toggleChecked(RecyclerAdapter.java:179)
at com.amaze.filemanager.ui.fragments.MainFragment.onListItemClicked(MainFragment.java:799)
at com.amaze.filemanager.adapters.RecyclerAdapter.lambda$onBindViewHolder$0(RecyclerAdapter.java:552)
at com.amaze.filemanager.adapters.RecyclerAdapter.lambda$onBindViewHolder$0$RecyclerAdapter(Unknown Source:0)
at com.amaze.filemanager.adapters.-$$Lambda$RecyclerAdapter$FDaxlO2LtK_Q0Z-YEozjkaXX_T8.onClick(Unknown Source:11)
at android.view.View.performClick(View.java:6294)
at android.view.View$PerformClick.run(View.java:24774)
at android.os.Handler.handleCallback(Handler.java:790)
at android.os.Handler.dispatchMessage(Handler.java:99)
at android.os.Looper.loop(Looper.java:164)
at android.app.ActivityThread.main(ActivityThread.java:6518)
at java.lang.reflect.Method.invoke(Native Method)
at com.android.internal.os.RuntimeInit$MethodAndArgsCaller.run(RuntimeInit.java:438)
at com.android.internal.os.ZygoteInit.main(ZygoteInit.java:807)`

/*
https://stackoverflow.com/a/37335777/3124150
*/
func remove(slice []string, s int) []string {
    return append(slice[:s], slice[s+1:]...)
}

/*
https://stackoverflow.com/a/56129336/3124150
*/
// NOTE: this isn't multi-Unicode-codepoint aware, like specifying skintone or
//       gender of an emoji: https://unicode.org/emoji/charts/full-emoji-modifiers.html
func substr(input string, start int, length int) string {
    asRunes := []rune(input)
    
    if start >= len(asRunes) {
        return ""
    }
    
    if start+length > len(asRunes) {
        length = len(asRunes) - start
    }
    
    return string(asRunes[start : start+length])
}

/*
https://stackoverflow.com/a/13582881/3124150
*/
func hash(s string) uint32 {//TODO check hash function
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}


func logsDistance(a string, b string) bool {
	a = strings.ReplaceAll(a, "\\n", "\n")
	b = strings.ReplaceAll(b, "\\n", "\n")
	
	
	a = strings.ReplaceAll(a, "\n", "")
	b = strings.ReplaceAll(b, "\n", "")
	
	
	if !strings.Contains(a, "com.amaze.filemanager") && 
		!strings.Contains(b, "com.amaze.filemanager") {
		return false //probably cant be fixed and checking if two issues are the same is complicated
	}
	
	
	if strings.Contains(a, "com.amaze.filemanager") != strings.Contains(b, "com.amaze.filemanager") {
		return false
	}
	
	
	splitA := strings.Split(a, "at")
	splitB := strings.Split(b, "at")
	
	
	for i := 0; i < len(splitA); i++ {
		if !strings.Contains(splitA[i], "com.amaze.filemanager") {
			splitA = remove(splitA, i)
			i--
		}
	}
	
	for i := 0; i < len(splitB); i++ {
		if !strings.Contains(splitB[i], "com.amaze.filemanager") {
			splitB = remove(splitB, i)
			i--
		}
	}
	

	for i := 0; i < len(splitA); i++ {
		if strings.Contains(splitA[i], "$") {
			splitA = remove(splitA, i)
			i--
		}
	}
	
	for i := 0; i < len(splitB); i++ {
		if strings.Contains(splitB[i], "$") {
			splitB = remove(splitB, i)
			i--
		}
	}
	
	
	if len(splitA) != len(splitB) {
		return false
	}
	
	
	equal := true
	for i := 0; i < len(splitA); i++ {
		startIndexA := strings.Index(splitA[i], "com")
		endIndexA := strings.Index(splitA[i], ":")
		startIndexB := strings.Index(splitB[i], "com")
		endIndexB := strings.Index(splitB[i], ":")
		
		
		if !(startIndexA == startIndexB && endIndexA == endIndexB) {
			equal = equal || false
			break
		}
		
		fmt.Sprintf("%s", splitA[i])
		fmt.Sprintf("%s", splitB[i])
		
		//Used hash here so this function is easily converted to hash of a single string
		hashA := hash(substr(splitA[i], startIndexA, endIndexA-startIndexA))
		hashB := hash(substr(splitB[i], startIndexB, endIndexB-startIndexB))
		
		if hashA != hashB {
			equal = equal || false
			break
		}
	}
	
	
	return equal
}

func TestComparator(t *testing.T) {
    var tests = []struct {
        a, b string
        want bool
    }{
        {test1A, test1B, true},
    }

    for _, tt := range tests {

        testname := fmt.Sprintf("%s,%s", tt.a, tt.b)
        t.Run(testname, func(t *testing.T) {
            ans := logsDistance(tt.a, tt.b)
            if ans != tt.want {
                t.Errorf("got %t, want %t", ans, tt.want)
            }
        })
    }
}
