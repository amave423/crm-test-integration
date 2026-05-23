import React from 'react';
import PodelitsiIcon from '../../assets/Podelitsi.svg';
import CopySubIcon from '../../assets/copy_sub.svg';

export default function ShareLinkBox({ link }) {
    const handleCopy = () => {
        if (link) navigator.clipboard.writeText(link);
    };

    return (
        <div className="share-link-box">
            <div className="share-link-box-inner">
                <img src={PodelitsiIcon} alt="Ссылка" />
                <p>Ссылка на тест</p>
            </div>
            <div className="share-link-box-inner1">
                <div className="share-link-input-box">
                    <input type="text" value={link || 'Появится после сохранения'} readOnly />
                    <button onClick={handleCopy} disabled={!link}>
                        <img src={CopySubIcon} alt="Копировать" className="copy-sub-icon" />
                    </button>
                </div>
            </div>
        </div>
    );
}
