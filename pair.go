// pair.go
//
// Aimed to manage Interface/Peer pairs

package main

// Pair is merely a structure which links
// a node and its public address
// type Pair struct {
// 	name      string
// 	publicKey string
// }

// func getPublicKeyFromFile(file string) (string, error) {
// 	vpn, err := ReadVPN(file)
// 	if err != nil {
// 		return "", fmt.Errorf("Error while retrieving public key from %s (%v)", file, err)
// 	}
// 	return vpn.server.Public(), nil
// }

// func extractPairsFromFolder(folder string) []Pair {
// 	files, err := ioutil.ReadDir(folder)
// 	if err != nil {
// 		return nil
// 	}
// 	pairs := make([]Pair, 0)
// 	for _, f := range files {
// 		name := strings.TrimSuffix(f.Name(), DefaultConfigSuffix)
// 		if !f.IsDir() {
// 			if pk, err := getPublicKeyFromFile(path.Join(folder, f.Name())); err == nil {
// 				pairs = append(pairs, Pair{name: name, publicKey: pk})
// 			}
// 		}
// 	}
// 	return pairs
// }

// func checkPeersAndClients(serverConfig string, clientDir string) error {
// 	serverKey, err := getPublicKeyFromFile(serverConfig)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
