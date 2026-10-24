package main

import (
	"bufio"
	"cmp"
	"fmt"
	"os"
	"strconv"
	"strings"
)


type Node[K cmp.Ordered] struct{
	Key K
	Value any
	Left *Node[K] 
	Right *Node[K]
}

type Tree[K cmp.Ordered] struct{
	Root *Node[K]
	Nodes []*Node[K]
}

func (tree *Tree[K]) Insert(key K, value any) {
	if tree.Root == nil {
		tree.Root =  &Node[K]{Key: key, Value: value}
		tree.Nodes = append(tree.Nodes, tree.Root) 
	} else {
		tree.InsertNode(tree.Root, key, value)
	}
			
} 

func (tree *Tree[K]) InsertNode(node *Node[K], key K, value any){
	if key == node.Key {
		node.Value = value
		return
	}

	if key < node.Key  {
		if node.Left == nil {
			node.Left = &Node[K]{Key: key, Value: value}
			tree.Nodes = append(tree.Nodes, node.Left)     
			return
		}
		tree.InsertNode(node.Left, key, value)
	} else {
		if node.Right == nil {
			node.Right = &Node[K]{Key: key, Value: value}
			tree.Nodes = append(tree.Nodes, node.Right) 
			return
		}
		tree.InsertNode(node.Right, key, value)
	}
}


func (tree *Tree[K]) Find(key K) any {
	return tree.FindByNode(tree.Root, key)
}

func (tree *Tree[K]) FindByNode(node *Node[K], key K) any {
	if node == nil {
		return "not found"
	}
	
	if node.Key == key{
		return node.Value
	}

	if key < node.Key {
		return tree.FindByNode(node.Left, key)
	}

	return tree.FindByNode(node.Right, key)
}

func (tree *Tree[K]) Delete(key K){
	if tree.Root == nil {
		return
	}
	root := tree.Root
	if root.Key == key {
		 tree.DeleteByNode(nil, tree.Root, key)
	} else {
		child := root.Left
		if key > root.Key{
			child = root.Right
		} 
		tree.DeleteByNode(root, child, key)
	}
}


func (tree *Tree[K]) DeleteByNode(parent *Node[K], child *Node[K], key K){
	if child == nil {
		return
	}

	if child.Key == key {
		//удаляем
		//1 случай - это листок, делаем его nil
		//2 - один потомок - нашли перезаписываем left или right предка на этого потомка   
		//3 - два потомка - находим самый левый из правых потомков, left его предка делаем nil,
		// left и right найденного = left и right удаляемого  
		if child.Left == nil && child.Right == nil {
			if parent != nil {
				if child.Key > parent.Key {
					parent.Right = nil
				} else {
					parent.Left = nil
				}
			} else {
				tree.Root = nil
			}
			return
		}
		//2 случай - есть один предок

		if child.Left == nil || child.Right == nil {
			var grandchild *Node[K]
			if child.Left == nil {
				grandchild = child.Right
			} else {
				grandchild = child.Left
			}
			
			if parent == nil {
				tree.Root = grandchild
				return
			}

			if parent.Left == child{
				parent.Left = grandchild
			} else {
				parent.Right = grandchild
			}
			return
		}

		//3 случай - два потомка

		//ищем минимальный узел правого поддерева
		var atLeastOnce bool
		targetNode := child.Right
		parentOfTargetNode := child
		for {
			if targetNode.Left != nil {
				parentOfTargetNode, targetNode = targetNode, targetNode.Left
				atLeastOnce = true
				continue
			}
			break
		}

		child.Key = targetNode.Key
		child.Value = targetNode.Value

		if atLeastOnce {
			parentOfTargetNode.Left = targetNode.Right
		} else {
			if parentOfTargetNode == child {
				parentOfTargetNode.Right = targetNode.Right   
			} else {
				parentOfTargetNode.Left = targetNode.Right
			}
		}

	} else {
		//ищем дальше
		if key < child.Key{
			if child.Left != nil {
				tree.DeleteByNode(child, child.Left, key)
			}		
		} else {
			if child.Right != nil {
				tree.DeleteByNode(child, child.Right, key)
			}
		}
	}
}

func (tree *Tree[K])InOrder(){
	sb := &strings.Builder{}
	if tree.Root != nil {
		tree.BuildOrder(tree.Root, sb)
	}
	trimmedString := strings.TrimSpace(sb.String())
	if trimmedString != "" {
		fmt.Println(trimmedString)
	}
}

func (tree *Tree[K])BuildOrder(node *Node[K], sb *strings.Builder) *strings.Builder{
	if node.Left != nil {
		tree.BuildOrder(node.Left, sb)	
	}
	fmt.Fprint(sb, node.Value, " ")
	if node.Right != nil {
		tree.BuildOrder(node.Right, sb)
	}
	return sb
}

func (tree *Tree[K])PrintTree() {
	for _, node := range tree.Nodes {
		fmt.Println(node.PrintNode())
	}
}

func (node *Node[K]) PrintNode() string {
	key := node.Key
	value := node.Value
	var right, left string
	if node.Right == nil {
		right = "nil"
	} else {
		right = fmt.Sprint(node.Right.Key)
	}
	if node.Left == nil {
		left = "nil"
	} else {
		left = fmt.Sprint(node.Left.Key)
	}
	result := fmt.Sprint(" Node Key: ",key,
					" Node Value: ", value,
					" Node Right: ", right,
					" Node Left: ", left)
	return result
}


func main(){
	scanner := bufio.NewScanner(os.Stdin)
	tree := Tree[int]{}

	
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(strings.TrimSpace(line), " ")
		var key int
		var err error
		if len(words) > 1 {
			key, err = strconv.Atoi(words[1])
			if err!= nil{
				fmt.Println(err)
			}
		}

		switch words[0] {
		case "Insert":
			value := words[2]
			tree.Insert(key, value)	
		case "Find":
			fmt.Println(tree.Find(key))
		case "Delete":
			tree.Delete(key)
		case "InOrder":	
			tree.InOrder()
		case "Exit":
			os.Exit(0)
		case "Tree":
			tree.PrintTree() 
		}
	}

	if scanner.Err() != nil{
		fmt.Println(scanner.Err().Error())
	}
	
}