using System.Collections;
using System.Collections.Generic;
using UnityEngine;

using DG.Tweening;

public class EnergyController : MonoBehaviour
{

    private GameObject mainCamera = null;

    // Start is called before the first frame update
    void Start()
    {
        mainCamera = GameObject.FindWithTag("MainCamera");
        transform.DOLocalRotate(transform.forward * 360f * 3, 3.0f, RotateMode.FastBeyond360).SetEase(Ease.Linear).SetLoops(-1, LoopType.Restart);
    }

    // Update is called once per frame
    void Update()
    {
        transform.LookAt(mainCamera.transform);
    }
}
