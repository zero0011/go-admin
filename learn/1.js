/**
 * @param {number[]} nums1
 * @param {number} m
 * @param {number[]} nums2
 * @param {number} n
 * @return {void} Do not return anything, modify nums1 in-place instead.
 */

const nums1 = [1,3,5]
const nums2 = [2,4,6]



var merge = function(nums1, m, nums2, n) {
    // 双指针解法
    let i = m - 1;
    let j = n - 1;
    let k = m + n - 1;
    while(j >= 0) {
        if (i < 0 || nums2[j] > nums1[i]) {
            nums1[k--] = nums2[j--];
        } else {
            nums1[k--] = nums1[i--];
        }
    }


};