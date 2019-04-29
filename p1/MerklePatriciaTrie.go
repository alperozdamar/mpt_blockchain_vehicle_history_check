package p1

import (
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strings"
)

const NoChildFound = uint8(88)
const LOGGING_LEVEL = "ERROR"

//sample input ="ab" output=[6,1,6,2]
func ConvertStringToHexArray(s string) []uint8 {
	hexArray := []uint8(s)
	convertedHexArray := []uint8{}
	for _, element := range hexArray {
		asInt := uint8(element) //int8 is the set of all signed 8-bit integers. Range: -128 through 127.
		//	fmt.Println("Ascii:", asInt) //Ascii value of a->97
		firstPart := asInt / 16
		secondPart := asInt % 16
		//	fmt.Println("firstPart:", firstPart)
		//	fmt.Println("secondPart:", secondPart)
		convertedHexArray = append(convertedHexArray, firstPart)
		convertedHexArray = append(convertedHexArray, secondPart)
	}
	return convertedHexArray
}

//!!!flag valuable applicable for extension and leaff
type Flag_value struct {
	encoded_prefix []uint8 //Path(Hex)
	value          string  //if leaf value:apple or if it's extension it is the pointer of next node(hashValue)
}

type Node struct {
	node_type    int        // 0: Null, 1: Branch, 2: Ext or Leaf
	branch_value [17]string // Ex: if its Branche and the hash is f, we should look 15. index of this array. So go to that node.
	//if  f is null. That means that the path is not exist in tree. You should add it !
	flag_value Flag_value //if node is ext or leaf we should look for the Flag_value
}

type MerklePatriciaTrie struct {
	DB   map[string]Node //key:Hash_node of Node, //value : Node's itself
	Root string          //Hash_value of Root
	//For example : When I give Root as key to the map, it will return actual Root Node.
}

func CompareCurrentPathAndKeyPathReturnMatchedIndexAndRemainingKeyPath(currentNodePath, key []uint8) (int, []uint8) {
	var matchIndex int
	var remainingKey []uint8
	matchIndex = -1
	if len(currentNodePath) == 0 || len(key) == 0 {
		if LOGGING_LEVEL == "DEBUG" {
			fmt.Println("WARNING: currentNodePath OR key length is 0.")
		}
		return matchIndex, remainingKey
	}
	if len(currentNodePath) > len(key) {
		if LOGGING_LEVEL == "DEBUG" {
			fmt.Println("CurrentNodePath is bigger than key. Return matchIndex:", matchIndex, " and remaining Key:", remainingKey)
		}
		return matchIndex, remainingKey
	}
	for i, value := range currentNodePath {
		if value == key[i] {
			matchIndex++
		} else {
			break
		}
	}
	if matchIndex == len(key) {
		return matchIndex, remainingKey
	}
	remainingKey = key[matchIndex+1:]

	if LOGGING_LEVEL == "DEBUG" {
		fmt.Println("Return matchIndex:", matchIndex, " and remaining Key:", remainingKey)
	}
	return matchIndex, remainingKey
}

/**
Description: The Get function takes a key as argument, traverses down the Merkle Patricia Trie
to find the value, and returns it. If the key doesn't exist, it will return an empty string.
(for the Go version: if the key is nil, Get returns an empty string.)
Arguments: key (string)
Return: the value stored for that key (string).
*/
func (mpt *MerklePatriciaTrie) Get(key string) (string, error) {
	var nodeType uint8
	var value string
	var errorMsg error
	nodeType = 0
	if key == "" {
		fmt.Println("ERROR: You can not enter empty String as a key.")
		return "", errors.New("Error: Invalid KEY!")
	}
	currentNode := mpt.DB[mpt.Root]
	if currentNode.node_type == 2 {
		nodeType = DetermineNodeType(currentNode)
	}
	keyPath := ConvertStringToHexArray(key)
	//if leaf
	// Compare convertedhexArray with Decoded node.flag_value.encoded_prefix return value
	//																				return error Not Found
	//if extension
	// 	if path is same Compare convertedhexArray with Decoded node.flag_value.encoded_prefix GO DOWN !!!!
	//										Check differentValue in node.branchValue[differentValue]
	for len(keyPath) != 0 && value == "" && errorMsg == nil {
		// if node_type is NULL, Branch, Leaf/Ext
		if nodeType < 2 {
			/**********************
			* EXTENSION OR BRANCHE !!!!!
			********************/
			valueReturn, errorMessageReturn, remainingKey, nodeTypeReturn, nextNode := mpt.GetLeafNode(currentNode, keyPath)
			value = valueReturn
			errorMsg = errorMessageReturn
			keyPath = remainingKey
			nodeType = nodeTypeReturn
			currentNode = nextNode
		} else {
			/**********************
			* LEAF	!!!!!
			********************/
			valueReturn, errorMessageReturn, remainingKey := GetLeafValue(currentNode, keyPath)
			if LOGGING_LEVEL == "DEBUG" {
				println("This is a LEAF!")
				fmt.Println("remainingKey:", remainingKey)
			}
			value = valueReturn
			errorMsg = errorMessageReturn
		}
	}
	return value, errorMsg
}

func DetermineNodeType(node Node) uint8 {
	var nodeType uint8
	encodedPrefix := node.flag_value.encoded_prefix[0]
	nodeType = encodedPrefix / 16
	return nodeType
}

func (mpt *MerklePatriciaTrie) AllMatchedInCurrentNode(node Node) (string, error, []uint8, uint8, Node) {
	var value string
	var nextNode Node
	if node.node_type == 1 {
		// All Matched, Extention--->Branch
		value = node.branch_value[16]
		return value, nil, nil, 10, node
	} else if node.node_type == 2 {
		// All Matched, Extention--->Leaf
		nodeType := DetermineNodeType(node)
		if nodeType == 2 || nodeType == 3 {
			value = node.flag_value.value
			return value, nil, nil, nodeType, node
		} else if nodeType == 0 || nodeType == 1 {
			nextNode := mpt.DB[node.flag_value.value]
			if nextNode.node_type == 1 {
				value = nextNode.branch_value[16]
				return value, nil, nil, nodeType, node
			} else if nextNode.node_type == 2 {
				// LEAF!
				value = nextNode.flag_value.value
				return value, nil, nil, nodeType, node
			}
		} else {
			return value, errors.New("Invalid_node_type"), nil, 10, nextNode
		}
	} else {
		return value, errors.New("Invalid_node_type"), nil, 10, nextNode
	}
	return value, errors.New("Invalid_node_type"), nil, 10, nextNode
}

func (mpt *MerklePatriciaTrie) PartiallyMatchFromExtension(node Node, partialKey []uint8, nodeType uint8) (string, error, []uint8, uint8, Node) {
	var value string
	nextNode := mpt.DB[node.flag_value.value]
	if nextNode.node_type == 1 && nextNode.branch_value[partialKey[0]] != "" {
		// PartialMatch, Extention--->Branche
		nextNode = mpt.DB[nextNode.branch_value[partialKey[0]]]
		if len(partialKey) == 1 {
			if nextNode.node_type == 1 {
				value = nextNode.branch_value[16]
				return value, nil, nil, nodeType, nextNode
			} else {
				value = nextNode.flag_value.value
				return value, nil, nil, nodeType, nextNode
			}
		} else {
			partialKey = partialKey[1:]
			return "", nil, partialKey, nodeType, nextNode
		}
	} else if nextNode.node_type == 2 {
		// PartialMatch, Extention--->Leaf
		nodeType = DetermineNodeType(nextNode)
	} else {
		return "", errors.New("Invalid_Node_Type"), partialKey, nodeType, nextNode
	}
	return value, nil, partialKey, nodeType, nextNode
}

func (mpt *MerklePatriciaTrie) PartiallyMatchFromBranch(node Node, partialKey []uint8, keyPath []uint8, nodeType uint8) (string, error, []uint8, uint8, Node) {
	var value string
	//Current Node Branch
	nextNode := mpt.DB[node.branch_value[keyPath[0]]]
	// PartialMatch,Branch--->Branch
	if nextNode.node_type == 1 {
		nextNode = mpt.DB[nextNode.branch_value[partialKey[0]]]
		if len(partialKey) == 1 {
			if nextNode.node_type == 1 {
				value = nextNode.branch_value[16]
				return value, nil, nil, nodeType, nextNode
			} else {
				value = nextNode.flag_value.value
				return value, nil, nil, nodeType, nextNode
			}
		} else {
			partialKey = partialKey[1:]
			return "", nil, partialKey, nodeType, nextNode
		}
	} else if nextNode.node_type == 2 {
		/********
		// Partially Match, Branch--->Extention
		// Partially Match, Branch--->Leaf
		******/
		nodeType = DetermineNodeType(nextNode)
		if nodeType > 1 {
			//Leaf
			value = nextNode.flag_value.value
			return value, nil, nil, nodeType, nextNode
		} else {
			//Extension
			nextNode = mpt.DB[nextNode.flag_value.value]
			partialKey = partialKey[1:]
			return "", nil, partialKey, nodeType, nextNode
		}
	} else {
		return "", errors.New("Invalid_node_type"), partialKey, nodeType, nextNode
	}
}

func (mpt *MerklePatriciaTrie) AllMatchFromBranch(node Node, keyPath []uint8, nodeType uint8) (string, error, []uint8, uint8, Node) {
	var value string
	//Current node branche
	nextNode := mpt.DB[node.branch_value[keyPath[0]]]
	if nextNode.node_type == 1 {
		// All Matched, Branch--->Leaf
		value = nextNode.branch_value[16]
		return value, nil, nil, 10, nextNode
	} else if nextNode.node_type == 2 {
		// All Matched, Branch--->Leaf
		nodeType = DetermineNodeType(nextNode)
		if nodeType == 2 || nodeType == 3 {
			value = nextNode.flag_value.value
			return value, nil, nil, nodeType, nextNode
		} else {
			return "", errors.New("Invalid_node_type"), nil, 10, nextNode
		}
	} else {
		return "", errors.New("Invalid_node_type"), nil, 10, nextNode
	}
}

func (mpt *MerklePatriciaTrie) GetLeafNode(node Node, keyPath []uint8) (string, error, []uint8, uint8, Node) {
	var partialKey []uint8
	var nodeType uint8
	var nextNode Node
	var currentKey []uint8

	// 2:Extension or Leaf
	if node.node_type == 2 {
		if LOGGING_LEVEL == "DEBUG" {
			fmt.Println("*******This is an Extension or Leaf*********")
		}
		currentKey = compact_decode(node.flag_value.encoded_prefix)
		nodeType = DetermineNodeType(node)
	} else if node.node_type == 1 {
		// 1:Branch
		if LOGGING_LEVEL == "DEBUG" {
			fmt.Println("*******This is a Branch*********")
		}
		currentKey = calculateCurrentKey(currentKey, node)
	} else {
		// 0:Null
		if LOGGING_LEVEL == "DEBUG" {
			fmt.Println("**********Node is NULL*************")
		}
		return "", errors.New("path_not_found"), partialKey, nodeType, nextNode
	}
	matchedIndex, partialKey := CompareCurrentPathAndKeyPathReturnMatchedIndexAndRemainingKeyPath(currentKey, keyPath)
	if node.node_type == 2 && matchedIndex+1 == len(currentKey) && len(partialKey) == 0 {
		return mpt.AllMatchedInCurrentNode(node)
	}
	if node.node_type == 1 && node.branch_value[keyPath[0]] != "" && len(partialKey) == 0 {
		return mpt.AllMatchFromBranch(node, keyPath, nodeType)
	}
	if matchedIndex >= 0 || node.branch_value[keyPath[0]] != "" {
		if matchedIndex+1 == len(currentKey) && len(partialKey) != 0 && node.node_type == 2 {
			return mpt.PartiallyMatchFromExtension(node, partialKey, nodeType)
		}
		if node.branch_value[keyPath[0]] != "" && len(partialKey) != 0 && node.node_type == 1 {
			return mpt.PartiallyMatchFromBranch(node, partialKey, keyPath, nodeType)
		}
	} else {
		return "", errors.New("path_not_found"), partialKey, nodeType, nextNode
	}
	return "", nil, partialKey, nodeType, nextNode
}

func calculateCurrentKey(currentKey []uint8, node Node) []uint8 {
	branchCurrentKey := node.branch_value
	var index uint8
	for index = 0; index < uint8(len(branchCurrentKey)); index++ {
		if branchCurrentKey[index] != "" {
			currentKey = append(currentKey, index)
		}
	}
	return currentKey
}

func GetLeafValue(node Node, searchPath []uint8) (string, error, []uint8) {
	currentPath := compact_decode(node.flag_value.encoded_prefix)
	matchedIndex, remainPath := CompareCurrentPathAndKeyPathReturnMatchedIndexAndRemainingKeyPath(currentPath, searchPath)
	// If all matched, return Leaf Value
	if matchedIndex+1 == len(currentPath) {
		value := node.flag_value.value
		return value, nil, remainPath
	} else {
		return "", errors.New("path_not_found"), remainPath
	}
}

/******************************************************** INSERT *******************************************************************/

/*
Description: The Insert function takes a key and value as arguments. It will traverse  Merkle Patricia Trie,
find the right place to insert the value, and do the insertion.(for the Go version: you can assume the key
and value will never be nil.)
Arguments: key (string), value (string)
Return: string
*/
func (mpt *MerklePatriciaTrie) Insert(key string, nodeValue string) {
	keyPath := ConvertStringToHexArray(key)
	if len(keyPath) == 0 {
		if mpt.Root == "" {
			rootNode := CreateNewLeafNodeWithValue(keyPath, nodeValue)
			mpt.Root = rootNode.hash_node()
			mpt.DB[mpt.Root] = rootNode
		} else {
			rootNode := mpt.GetNodeByHashCode(mpt.Root)
			if rootNode.IsBranch() {
				delete(mpt.DB, mpt.Root)
				rootNode.branch_value[16] = nodeValue
				mpt.Root = rootNode.hash_node()
				mpt.DB[mpt.Root] = rootNode
			} else {
				newRootNode := mpt.MergeLE(rootNode, keyPath, nodeValue)
				mpt.Root = newRootNode.hash_node()
			}
		}
	} else if mpt.Root == "" {
		rootNode := CreateNewLeafNodeWithValue(keyPath, nodeValue)
		mpt.Root = rootNode.hash_node()
		mpt.DB[mpt.Root] = rootNode
	} else {
		nodePath, remainingPath := mpt.GetNodePath(keyPath, mpt.Root)
		if len(nodePath) == 0 {
			rootNode := mpt.GetNodeByHashCode(mpt.Root)
			if rootNode.IsBranch() {
				childNode := CreateNewLeafNodeWithValue(keyPath[1:], nodeValue)
				childHash := childNode.hash_node()
				mpt.DB[childHash] = childNode
				rootNode.branch_value[keyPath[0]] = childHash
				delete(mpt.DB, mpt.Root)
				mpt.Root = rootNode.hash_node()
				mpt.DB[mpt.Root] = rootNode
			} else {
				newRootNode := mpt.MergeLE(rootNode, remainingPath, nodeValue)
				mpt.Root = newRootNode.hash_node()
			}
		} else {
			lastPrefixNode := nodePath[len(nodePath)-1]
			if lastPrefixNode.IsBranch() {
				if len(remainingPath) == 0 {
					mpt.ModifyHashChain(nodePath, 16, nodeValue)
				} else if lastPrefixNode.branch_value[remainingPath[0]] == "" {
					newLeafNode := CreateNewLeafNodeWithValue(remainingPath[1:], nodeValue)
					newLeafNodeHash := newLeafNode.hash_node()
					mpt.DB[newLeafNodeHash] = newLeafNode
					mpt.ModifyHashChain(nodePath, remainingPath[0], newLeafNodeHash)
				} else {
					childNode := mpt.GetNodeByHashCode(lastPrefixNode.branch_value[remainingPath[0]])
					newChildNode := mpt.MergeLE(childNode, remainingPath[1:], nodeValue)
					mpt.ModifyHashChain(nodePath, remainingPath[0], newChildNode.hash_node())
				}
			} else {
				if lastPrefixNode.IsLeaf() == false {
					fmt.Println("The last node on a prefix path cannot be an extension: ", nodePath)
				}
				lastPrefixNodeHash := lastPrefixNode.hash_node()
				newLastPrefixNode := mpt.MergeLE(
					lastPrefixNode,
					append(compact_decode(lastPrefixNode.flag_value.encoded_prefix), remainingPath...),
					nodeValue)
				if len(nodePath) == 1 {
					mpt.Root = newLastPrefixNode.hash_node()
				} else {
					parentNode := nodePath[len(nodePath)-2]
					childIndex := parentNode.GetBranchChildIndex(lastPrefixNodeHash)
					mpt.ModifyHashChain(nodePath[:len(nodePath)-1], childIndex, newLastPrefixNode.hash_node())
				}
			}
		}
	}
}

func CreateNewLeafNodeWithValue(path []uint8, value string) Node {
	var node Node
	node.node_type = 2
	node.branch_value = [17]string{}
	var flagValue Flag_value
	flagValue.value = value
	flagValue.encoded_prefix = compact_encode(append(path, 16))
	node.flag_value = flagValue
	return node
}

func (node *Node) IsBranch() bool {
	return node.node_type == 1
}

/**
 * Recursive function to get Node Path.
 */
func (mpt *MerklePatriciaTrie) GetNodePath(keyPath []uint8, nodeHashCode string) ([]Node, []uint8) {
	node := mpt.GetNodeByHashCode(nodeHashCode)
	if node.IsLeaf() {
		nodePath := compact_decode(node.flag_value.encoded_prefix)
		if IsPathPrefix(nodePath, keyPath) {
			return []Node{node}, keyPath[len(nodePath):]
		}
		if LOGGING_LEVEL == "DEBUG" {
			fmt.Println("path:", keyPath)
		}
		return []Node{}, keyPath
	}
	if node.IsExtension() {
		extensionPath := compact_decode(node.flag_value.encoded_prefix)
		if !IsPathPrefix(extensionPath, keyPath) {
			return []Node{}, keyPath
		}
		recNodePath, recRemainingPath := mpt.GetNodePath(keyPath[len(extensionPath):], node.flag_value.value)
		if LOGGING_LEVEL == "DEBUG" {
			fmt.Println("recRemainingPath:", recRemainingPath)
			fmt.Println("recNodePath:", recNodePath)
		}
		return append([]Node{node}, recNodePath...), recRemainingPath
	}

	if (len(keyPath) == 0) || (node.branch_value[keyPath[0]] == "") {
		if LOGGING_LEVEL == "DEBUG" {
			fmt.Println("path:", keyPath)
		}
		return []Node{node}, keyPath
	}
	recursiveNodePath, recursiveRemainingPath := mpt.GetNodePath(keyPath[1:], node.branch_value[keyPath[0]])
	if len(recursiveNodePath) > 0 {
		if strings.EqualFold(LOGGING_LEVEL, "DEBUG") {
			fmt.Println("recursiveNodePath:", recursiveNodePath)
			fmt.Println("recursiveRemainingPath:", recursiveRemainingPath)
		}
		return append([]Node{node}, recursiveNodePath...), recursiveRemainingPath
	}
	if strings.EqualFold(LOGGING_LEVEL, "DEBUG") {
		fmt.Println("path:", keyPath)
	}
	return []Node{node}, keyPath
}

func CheckPathEquality(path1 []uint8, path2 []uint8) bool {
	if path1 == nil {
		return path2 == nil
	}
	if path2 == nil {
		return false
	}
	if len(path1) != len(path2) {
		return false
	}
	for i := 0; i < len(path1); i++ {
		if path1[i] != path2[i] {
			return false
		}
	}
	return true
}

func (node *Node) GetBranchChildIndex(childHash string) uint8 {
	for i := uint8(0); i < uint8(16); i++ {
		if node.branch_value[i] == childHash {
			return i
		}
	}
	return NoChildFound
}

func (node *Node) IsLeaf() bool {
	return (node.node_type == 2) && (node.flag_value.encoded_prefix[0]>>5 == 1)
}

func (node *Node) IsExtension() bool {
	//y >> z is "y divided by 2, z times".
	// >>   right shift integer >> unsigned integer
	return (node.node_type == 2) && (node.flag_value.encoded_prefix[0]>>5 == 0)
}

func IsPathPrefix(path1 []uint8, path2 []uint8) bool {
	return len(GetCommonPath(path1, path2)) == len(path1)
}

func GetCommonPath(path1 []uint8, path2 []uint8) []uint8 {
	length := len(path1)
	if length > len(path2) {
		length = len(path2)
	}
	commonPath := []uint8{}
	for i := 0; i < length; i++ {
		if path1[i] == path2[i] {
			commonPath = append(commonPath, path1[i])
		} else {
			break
		}
	}
	return commonPath
}

func (mpt *MerklePatriciaTrie) GetNodeByHashCode(hashCode string) Node {
	node, nodeExists := mpt.DB[hashCode]
	if nodeExists == false {
		fmt.Println("Could not find node with hashCode: ", hashCode)
	}
	return node
}

func CreateNewExtensionNodeWithHash(path []uint8, hash string) Node {
	var extensionNode Node
	extensionNode.node_type = 2
	extensionNode.branch_value = [17]string{}
	var flagValue Flag_value
	flagValue.encoded_prefix = compact_encode(path)
	flagValue.value = hash
	extensionNode.flag_value = flagValue
	return extensionNode
}

func CreateNewBranchNodeWithValue(value string) Node {
	branchValue := [17]string{}
	branchValue[16] = value
	var branchNode Node
	branchNode.node_type = 1
	branchNode.branch_value = branchValue
	branchNode.flag_value = Flag_value{}
	return branchNode
}

func (mpt *MerklePatriciaTrie) MergeLE(node Node, keyPath []uint8, new_value string) Node {
	nodePath := compact_decode(node.flag_value.encoded_prefix)
	delete(mpt.DB, node.hash_node())
	if CheckPathEquality(nodePath, keyPath) {
		if node.IsLeaf() {
			node.flag_value.value = new_value
			mpt.DB[node.hash_node()] = node
			return node
		}
		branchNode := mpt.GetNodeByHashCode(node.flag_value.value)
		branchNode.branch_value[16] = new_value
		branchNodeNewHash := branchNode.hash_node()
		mpt.DB[branchNodeNewHash] = branchNode
		node.flag_value.value = branchNodeNewHash
		mpt.DB[node.hash_node()] = node
		return node
	}
	commonPath := GetCommonPath(nodePath, keyPath)
	remainingNodePath := nodePath[len(commonPath):]
	remainingPath := keyPath[len(commonPath):]
	newBranchNode := CreateNewBranchNodeWithValue("")
	if node.IsLeaf() {
		if len(remainingNodePath) == 0 {
			newBranchNode.branch_value[16] = node.flag_value.value
		} else {
			newLeafNode := CreateNewLeafNodeWithValue(remainingNodePath[1:], node.flag_value.value)
			newLeafNodeHash := newLeafNode.hash_node()
			mpt.DB[newLeafNodeHash] = newLeafNode
			newBranchNode.branch_value[remainingNodePath[0]] = newLeafNodeHash
		}
	} else {
		if len(remainingNodePath) == 1 {
			newBranchNode.branch_value[remainingNodePath[0]] = node.flag_value.value
		} else {
			newExtensionNode := CreateNewExtensionNodeWithHash(remainingNodePath[1:], node.flag_value.value)
			newExtensionNodeHash := newExtensionNode.hash_node()
			mpt.DB[newExtensionNodeHash] = newExtensionNode
			newBranchNode.branch_value[remainingNodePath[0]] = newExtensionNodeHash
		}
	}
	if len(remainingPath) == 0 {
		newBranchNode.branch_value[16] = new_value
	} else {
		newLeafNode := CreateNewLeafNodeWithValue(remainingPath[1:], new_value)
		newLeafNodeHash := newLeafNode.hash_node()
		mpt.DB[newLeafNodeHash] = newLeafNode
		newBranchNode.branch_value[remainingPath[0]] = newLeafNodeHash
	}
	newBranchNodeHash := newBranchNode.hash_node()
	mpt.DB[newBranchNodeHash] = newBranchNode
	if len(commonPath) > 0 {
		newExtensionNode := CreateNewExtensionNodeWithHash(commonPath, newBranchNodeHash)
		mpt.DB[newExtensionNode.hash_node()] = newExtensionNode
		return newExtensionNode
	} else {
		return newBranchNode
	}
}

/******************************************************** INSERT SON*******************************************************************/

/******************************************************** DELETE *******************************************************************/

/*
Description: The Delete function takes a key as argument, traverses the Merkle Patricia Trie and finds that key.
If the key exists, delete the corresponding value and re-balance the trie if necessary,
	then return an empty string; if the key doesn't exist, return "path_not_found".
Arguments: key (string)
Return: string
*/
func (mpt *MerklePatriciaTrie) Delete(key string) (string, error) {
	path := ConvertStringToHexArray(key)
	nodePath, remainingPath := mpt.GetNodePath(path, mpt.Root)
	if (len(nodePath) == 0) || (len(remainingPath) > 0) {
		return "", errors.New("path_not_found")
	}
	lastPathNode := &nodePath[len(nodePath)-1]
	if lastPathNode.IsLeaf() {
		lastPathNodeHashCode := lastPathNode.hash_node()
		delete(mpt.DB, lastPathNodeHashCode)
		if len(nodePath) == 1 {
			mpt.Root = ""
		} else {
			parentNode := nodePath[len(nodePath)-2]
			mpt.ReOrderTreePath(nodePath[:len(nodePath)-1],
				parentNode.GetBranchChildIndex(lastPathNodeHashCode))
		}
		return "", nil
	}
	if lastPathNode.branch_value[16] == "" {
		return "", errors.New("path_not_found")
	}
	mpt.ReOrderTreePath(nodePath, 16)
	return "", nil
}

// Firstly we should calculate how many elements (children + value) the branch node has.
// If branche node array has at least 2 elements (including last one[16]),
// then we don't need to change the structure of the tree.
// Else, we might need to merge branch node with its parent and/or its child.
func (mpt *MerklePatriciaTrie) ReOrderTreePath(nodeChain []Node, branchToRemove uint8) {
	if len(nodeChain) == 0 {
		return
	}
	lastNode := nodeChain[len(nodeChain)-1]
	brancheElementCount := 0
	remainingChildIndex := NoChildFound
	for i := uint8(0); i < uint8(len(lastNode.branch_value)); i++ {
		if (i != branchToRemove) && (len(lastNode.branch_value[i]) > 0) {
			remainingChildIndex = i
			brancheElementCount++
		}
	}
	if brancheElementCount > 1 {
		mpt.ModifyHashChain(nodeChain, branchToRemove, "")
	} else {
		delete(mpt.DB, lastNode.hash_node())
		var modifiedNode Node
		if remainingChildIndex == 16 {
			modifiedNode = CreateNewLeafNodeWithValue([]uint8{}, lastNode.branch_value[16])
		} else {
			modifiedNode = CreateNewExtensionNodeWithHash(
				[]uint8{remainingChildIndex}, lastNode.branch_value[remainingChildIndex])
		}
		if modifiedNode.IsExtension() {
			childNode := mpt.GetNodeByHashCode(modifiedNode.flag_value.value)
			if childNode.IsExtension() || childNode.IsLeaf() {
				modifiedNode = modifiedNode.MergeExtensionWithExtensionOrLeaf(childNode)
				delete(mpt.DB, childNode.hash_node())
			}
		}
		if len(nodeChain) == 1 {
			mpt.Root = modifiedNode.hash_node()
			mpt.DB[mpt.Root] = modifiedNode
		} else {
			parentNode := nodeChain[len(nodeChain)-2]
			if parentNode.IsBranch() {
				modifiedNodeHash := modifiedNode.hash_node()
				mpt.DB[modifiedNodeHash] = modifiedNode
				mpt.ModifyHashChain(nodeChain[:len(nodeChain)-1], parentNode.GetBranchChildIndex(lastNode.hash_node()), modifiedNodeHash)
			} else {
				parentNodeHash := parentNode.hash_node()
				delete(mpt.DB, parentNodeHash)
				modifiedNode = parentNode.MergeExtensionWithExtensionOrLeaf(modifiedNode)
				modifiedNodeHash := modifiedNode.hash_node()
				mpt.DB[modifiedNodeHash] = modifiedNode
				if len(nodeChain) == 2 {
					mpt.Root = modifiedNodeHash
				} else {
					grandParentNode := nodeChain[len(nodeChain)-3]
					mpt.ModifyHashChain(
						nodeChain[:len(nodeChain)-2],
						grandParentNode.GetBranchChildIndex(parentNodeHash),
						modifiedNodeHash)
				}
			}
		}
	}
}

func (mpt *MerklePatriciaTrie) ModifyHashChain(hashChain []Node, child uint8, newNodeValue string) {
	if len(hashChain) == 0 {
		return
	}
	childArray := make([]uint8, len(hashChain))
	for i := 0; i < len(hashChain)-1; i++ {
		childArray[i] = NoChildFound
		if hashChain[i].IsBranch() {
			nextNodeHash := hashChain[i+1].hash_node()
			childArray[i] = hashChain[i].GetBranchChildIndex(nextNodeHash)
		}
	}
	childArray[len(childArray)-1] = child
	newValue := newNodeValue
	for i := len(hashChain) - 1; i >= 0; i-- {
		node := hashChain[i]
		delete(mpt.DB, node.hash_node())
		if node.IsLeaf() || node.IsExtension() {
			node.flag_value.value = newValue
		} else {
			node.branch_value[childArray[i]] = newValue
		}
		newValue = node.hash_node()
		mpt.DB[newValue] = node
	}
	mpt.Root = newValue
}

func (node *Node) MergeExtensionWithExtensionOrLeaf(child Node) Node {
	if child.IsExtension() {
		return CreateNewExtensionNodeWithHash(
			append(compact_decode(node.flag_value.encoded_prefix),
				compact_decode(child.flag_value.encoded_prefix)...),
			child.flag_value.value)
	}
	return CreateNewLeafNodeWithValue(
		append(compact_decode(node.flag_value.encoded_prefix),
			compact_decode(child.flag_value.encoded_prefix)...),
		child.flag_value.value)
}

////////////////// DELETE END /////////////////////////////////////////////////////////////

//Compact_encode() receives a "hex_data", add a prefix, then convert "prefix+hex_data" to its ASCII format.
//Compact_encode() is used when you need to create a Leaf or Ext Node. It is different from
//"converting key to ASCII then hex_data".
func compact_encode(hex_array []uint8) []uint8 {
	size := len(hex_array)
	term := 0
	if hex_array[size-1] == 16 {
		term = 1
	} else {
		term = 0
	}
	//fmt.Println("\nTerm:", term)
	//Remove 16 if Term==1 (If Term is 1,It is a leaf!)
	if term == 1 {
		hex_array = hex_array[:size-1]
	}
	size = len(hex_array)
	oddlen := size % 2
	flags := 2*term + oddlen //prefix
	if oddlen == 1 {
		//hexarray = [flags] + hex_array
		hex_array = append([]uint8{uint8(flags)}, hex_array...)
	} else {
		//hexarray = [flags] + [0] + hex_array
		hex_array = append([]uint8{0}, hex_array...)
		hex_array = append([]uint8{uint8(flags)}, hex_array...)
		// hexarray now has an even length whose first nibble is the flags.
	}
	//ab=>006162 => 09798
	//kl=>00611612 => 00107108
	var encodedArray []uint8
	for i := 0; i < len(hex_array); i += 2 {
		asciiValue := (16*hex_array[i] + 1*hex_array[i+1])
		//fmt.Println("asciiValue:", asciiValue)
		encodedArray = append(encodedArray, asciiValue)
	}
	return encodedArray
}

// If Leaf, ignore 16 at the end
func compact_decode(encoded_arr []uint8) []uint8 {
	var decodedArray []uint8
	for i := 0; i < len(encoded_arr); i += 1 {
		//fmt.Print("\nHex-Value:", encoded_arr[i])
		hexValueFirstPart := encoded_arr[i] / 16
		hexValueSecondPart := encoded_arr[i] % 16
		decodedArray = append(decodedArray, hexValueFirstPart)
		decodedArray = append(decodedArray, hexValueSecondPart)
	}
	//0 Extension node, even number of nibbles
	//2 Leaf node, even number of nibbles
	if decodedArray[0] == 0 || decodedArray[0] == 2 {
		decodedArray = append(decodedArray[:0], decodedArray[1:]...)
		decodedArray = append(decodedArray[:0], decodedArray[1:]...)
	} else if decodedArray[0] == 1 || decodedArray[0] == 3 {
		//1 Extension node, odd number of nibbles
		//3 Leaf node, odd number of nibbles
		decodedArray = append(decodedArray[:0], decodedArray[1:]...)
	}
	//return []uint8{}
	return decodedArray
}

func (node *Node) hash_node() string {
	var str string
	switch node.node_type {
	case 0:
		str = ""
	case 1:
		str = "branch_"
		for _, v := range node.branch_value {
			str += v
		}
	case 2:
		str = node.flag_value.value
	}

	sum := sha3.Sum256([]byte(str))
	return "HashStart_" + hex.EncodeToString(sum[:]) + "_HashEnd"
}

func (node *Node) String() string {
	str := "empty string"
	switch node.node_type {
	case 0:
		str = "[Null Node]"
	case 1:
		str = "Branch["
		for i, v := range node.branch_value[:16] {
			str += fmt.Sprintf("%d=\"%s\", ", i, v)
		}
		str += fmt.Sprintf("value=%s]", node.branch_value[16])
	case 2:
		encoded_prefix := node.flag_value.encoded_prefix
		node_name := "Leaf"
		if is_ext_node(encoded_prefix) {
			node_name = "Ext"
		}
		ori_prefix := strings.Replace(fmt.Sprint(compact_decode(encoded_prefix)), " ", ", ", -1)
		str = fmt.Sprintf("%s<%v, value=\"%s\">", node_name, ori_prefix, node.flag_value.value)
	}
	return str
}

func node_to_string(node Node) string {
	return node.String()
}

func (mpt *MerklePatriciaTrie) Initial() {
	mpt.Root = ""
	mpt.DB = make(map[string]Node)
}

func is_ext_node(encoded_arr []uint8) bool {
	return encoded_arr[0]/16 < 2
}

func TestCompact() {
	Test_compact_encode()
}

func (mpt *MerklePatriciaTrie) String() string {
	content := fmt.Sprintf("ROOT=%s\n", mpt.Root)
	for hash := range mpt.DB {
		content += fmt.Sprintf("%s: %s\n", hash, node_to_string(mpt.DB[hash]))
	}
	return content
}

func (mpt *MerklePatriciaTrie) Order_nodes() string {
	raw_content := mpt.String()
	content := strings.Split(raw_content, "\n")
	root_hash := strings.Split(strings.Split(content[0], "HashStart")[1], "HashEnd")[0]
	queue := []string{root_hash}
	i := -1
	rs := ""
	cur_hash := ""
	for len(queue) != 0 {
		last_index := len(queue) - 1
		cur_hash, queue = queue[last_index], queue[:last_index]
		i += 1
		line := ""
		for _, each := range content {
			if strings.HasPrefix(each, "HashStart"+cur_hash+"HashEnd") {
				line = strings.Split(each, "HashEnd: ")[1]
				rs += each + "\n"
				rs = strings.Replace(rs, "HashStart"+cur_hash+"HashEnd", fmt.Sprintf("Hash%v", i), -1)
			}
		}
		temp2 := strings.Split(line, "HashStart")
		flag := true
		for _, each := range temp2 {
			if flag {
				flag = false
				continue
			}
			queue = append(queue, strings.Split(each, "HashEnd")[0])
		}
	}

	rs = strings.Replace(rs, "\n", "\r\n", -1)

	return rs
}

// For Project-2********************************************************************************************************
// For Traversing the MPT
func (mpt *MerklePatriciaTrie) MptTraverser(node Node, path []uint8, leafList map[string]string) {
	if node.IsBranch() {
		if node.branch_value[16] != "" {
			leafList[ConvertHexArrayToString(path)] = node.branch_value[16]
		}
		for i := uint8(0); i < uint8(16); i++ {
			if node.branch_value[i] != "" {
				mpt.MptTraverser(mpt.GetNodeByHashCode(node.branch_value[i]),
					append(path, i), leafList)
			}
		}
	} else if node.IsExtension() {
		mpt.MptTraverser(mpt.GetNodeByHashCode(node.flag_value.value),
			append(path, compact_decode(node.flag_value.encoded_prefix)...), leafList)
	} else {
		finalPath := append(path, compact_decode(node.flag_value.encoded_prefix)...)
		leafList[ConvertHexArrayToString(finalPath)] = node.flag_value.value
	}
}

func ConvertHexArrayToString(hexArray []uint8) string {
	stringByteArray := make([]byte, len(hexArray)/2)
	for i := 0; i < len(stringByteArray); i++ {
		stringByteArray[i] = hexArray[2*i]<<4 + hexArray[2*i+1]
	}
	return string(stringByteArray)
}

func (mpt *MerklePatriciaTrie) ToMptMap() map[string]string {
	leafList := make(map[string]string)
	if mpt.Root != "" {
		mpt.MptTraverser(mpt.GetNodeByHashCode(mpt.Root), make([]uint8, 0), leafList)
	}
	return leafList
}


//////////////// TEST METHODS ////////////////////////////////////////////////////////////////////////////

func Test_compact_encode() {
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{1, 2, 3, 4, 5})), []uint8{1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 1, 2, 3, 4, 5})), []uint8{0, 1, 2, 3, 4, 5}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{0, 15, 1, 12, 11, 8, 16})), []uint8{0, 15, 1, 12, 11, 8}))
	fmt.Println(reflect.DeepEqual(compact_decode(compact_encode([]uint8{15, 1, 12, 11, 8, 16})), []uint8{15, 1, 12, 11, 8}))
	fmt.Println(compact_decode(compact_encode([]uint8{5, 16})))
}
