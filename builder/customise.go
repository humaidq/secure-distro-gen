package builder

// customise modifies the distribution filesystem and configurations to match
// the configuration provided in the build session.
func customise(sess *buildSession) error {
	writeToFile(sess.extractDir+"/etc/issue", sess.cust.DistName+" \\r (\\n) (\\l)\n\n")
	return nil
}
