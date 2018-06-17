package main

import "crypto/sha256"

//Merkle 树的作用:一个节点可以在不下载整个块的情况下，验证是否包含某笔交易。
// 并且这些只需要一个交易哈希，一个 Merkle 树根哈希和一个 Merkle 路径。

//默克尔树
type MerkleTree struct {
	RootNode *MerkleNode
}

//节点
type MerkleNode struct {
	Left *MerkleNode
	Right *MerkleNode
	data []byte
}

//创建默克尔数
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode
	if len(data) % 2 != 0 {			//如果节点数为单数,复制最后一个节点数据并追加
		data = append(data, data[len(data) - 1])
	}
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}
	for i := 0; i < len(data)/2; i++ {		//一层层向上合并
		var newlevel []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j],&nodes[j+1],nil)
			newlevel = append(newlevel, *node)
		}
		nodes = newlevel
	}
	mTree := MerkleTree{&nodes[0]}
	return &mTree
}

//创建一个节点
func NewMerkleNode(left ,right *MerkleNode, data []byte) *MerkleNode {
	mNode := MerkleNode{}
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.data = hash[:]
	} else {
		preHashes := append(left.data, right.data...)
		hash := sha256.Sum256(preHashes)
		mNode.data = hash[:]
	}
	mNode.Left = left
	mNode.Right = right
	return &mNode
}
