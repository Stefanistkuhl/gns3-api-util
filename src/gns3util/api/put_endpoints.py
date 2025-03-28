from . import GNS3APIClient


class GNS3PutAPI(GNS3APIClient):

    # controller endpoint

    def iou_license(self, data):
        return self._api_call("iou_license", method="PUT", data=data)
